package service

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

func LoadQueriesFromLog(ctx context.Context, filePath string, format FormatType, cmd string) ([]string, error) {
	queries := []string{}

	filePath, err := getPath(filePath)
	if err != nil {
		return queries, ErrWrap(err, UserInputError)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return queries, ErrWrap(err, UserInputError)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		query, err := GetQueryByFormat(format, scanner.Text(), cmd)
		if err != nil {
			return queries, ErrWrap(err, UserInputError)
		}

		if query == "" {
			continue
		}
		queries = append(queries, query)
	}
	if err := scanner.Err(); err != nil {
		return queries, ErrWrap(err, UserInputError)
	}

	return queries, nil
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

func LoadQueriesFromDB(ctx context.Context) ([]string, error) {
	list := []string{}

	rows, err := query(officialDB, "select argument from general_log where command_type in ('Query', 'Execute')")
	if err != nil {
		return nil, ErrWrap(err, OtherError)
	}

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, ErrWrap(err, OtherError)
		}
		list = append(list, value)
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return list, ErrWrap(err, OtherError)
	}

	return list, nil
}
