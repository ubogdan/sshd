package client

import (
	"log"
)

// SetWinSize resizes the client to the given width and height
func (c *Client) SetWinSize(width, height int) {
	width = 1000 //todo:: work around
	err := c.term.SetSize(width, height)
	if err != nil {
		log.Println("SetWinSize failed: ", width, height)
	}
	c.termWidth, c.termHeight = width, height
}

func (c *Client) Resize(width, height int) {
	c.SetWinSize(width, height)
	if c.OnBastion {
		c.WinSizeChan <- WinSize{
			H: height,
			W: width,
		}
	}
}
