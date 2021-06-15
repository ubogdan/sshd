package client

import (
	"fmt"
	"log"
)

func (c *Client) writePlain(content string) {
	_, err := c.term.Write([]byte(content))
	if err != nil {
		log.Println(err)
	}
}

func (c *Client) restorePrompt() {
	prompt := fmt.Sprintf("❤ %s ❤", c.User.Name)

	if c.selectedGroup == nil && c.selectedFriend != nil {
		prompt = fmt.Sprintf("[%s -> %s]", c.User.Name, c.selectedFriend.Name)
	} else if c.selectedGroup != nil && c.selectedFriend == nil {
		prompt = fmt.Sprintf("[%s @ %s]", c.User.Name, c.selectedGroup.Name)
	}
	c.SetPrompt(prompt)
}
