package util

import (
	"errors"
	"fmt"
	"golang.org/x/term"
	"io"
	"log"
	"strconv"
	"strings"
)

func ListSelect(rows []fmt.Stringer, c io.ReadWriter, listPrompt, selectPrompt string, retry int) (one interface{}, err error) {
	terminal := term.NewTerminal(c, listPrompt)
	for idx, row := range rows {
		msg := fmt.Sprintf("%03d    %s\r\n", idx+1, row.String())
		_, err := terminal.Write([]byte(msg))
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < retry; i++ {
		selectPromptI := "Please input the row index you want:"
		terminal.SetPrompt(selectPromptI)
		line, err := terminal.ReadLine()
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "q" || line == "exit" || line == "quite" {
			return nil, errors.New("user exit the selection")
		}
		parseInt, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			safeWrite(terminal, []byte("your input is not a valid index number,please retry"))
			continue
		}
		if parseInt < 1 || int(parseInt) > len(rows) {
			safeWrite(terminal, []byte("your input number is out of range,please retry"))
			continue
		}
		return rows[int(parseInt-1)], nil
	}
	return nil, errors.New("sorry, you run out of retry")
}
func safeWrite(writer io.Writer, data []byte) {
	_, err := writer.Write(data)
	if err != nil {
		log.Println(err)
	}
}
