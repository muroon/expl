package expl

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
)

type FormatType string

const (
	FormatSimple   FormatType = "simple"
	FormatOfficial FormatType = "official"
	FormatCommand  FormatType = "command"
)

func LoadQueriesFromLogChannels(
	ctx context.Context, filePath string, format FormatType, cmd string,
) (<-chan string, <-chan error) {

	qCh := make(chan string)
	errCh := make(chan error)

	go func() {
		defer func() {
			close(qCh)
			close(errCh)
		}()

		filePath, err := getPath(filePath)
		if err != nil {
			errCh <- ErrWrap(err, UserInputError)
		}

		f, err := os.Open(filePath)
		if err != nil {
			errCh <- ErrWrap(err, UserInputError)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			query, err := GetQueryByFormat(format, scanner.Text(), cmd)
			if err != nil {
				errCh <- ErrWrap(err, UserInputError)
			}

			if query == "" {
				continue
			}
			qCh <- query
		}
		if err := scanner.Err(); err != nil {
			errCh <- ErrWrap(err, UserInputError)
		}
	}()

	return qCh, errCh
}

func LoadQueriesFromDBChannels(ctx context.Context) (<-chan string, <-chan error) {
	qCh := make(chan string)
	errCh := make(chan error)

	go func() {
		defer close(qCh)
		defer close(errCh)

		rows, err := query(officialDB, "select argument from general_log where command_type in ('Query', 'Execute')")
		if err != nil {
			errCh <- ErrWrap(err, OtherError)
			return
		}

		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				errCh <- ErrWrap(err, OtherError)
				return
			}
			qCh <- value
		}

		defer rows.Close()

		if err := rows.Err(); err != nil {
			errCh <- ErrWrap(err, OtherError)
			return
		}
	}()

	return qCh, errCh
}

func GetQueryByFormat(format FormatType, line, cmd string) (string, error) {
	query := ""

	if format == FormatOfficial {
		ws := strings.Split(line, "\t")

		if len(ws) == 3 {
			if strings.Index(ws[1], "Execute") > -1 || strings.Index(ws[1], "Query") > -1 {
				query = ws[2]
			}
		}
	} else if format == FormatCommand {

		c := exec.Command("sh", "-c", cmd)
		c.Stdin = strings.NewReader(line)

		var out bytes.Buffer
		c.Stdout = &out
		c.Start()
		c.Wait()

		query = out.String()

	} else {
		query = line
	}

	return query, nil
}
