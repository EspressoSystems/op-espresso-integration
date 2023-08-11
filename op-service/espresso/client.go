package espresso

type Client struct {
	url string
}

func NewClient(url string) (*Client, error) {
	c := new(Client)
	c.url = url
	return c, nil
}
