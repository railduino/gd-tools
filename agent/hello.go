package agent

func HelloTest(req *Request) bool {
	return req != nil && req.Hello != ""
}

func HelloHandler(req *Request, resp *Response) error {
	if req == nil || req.Hello == "" || resp == nil {
		return nil
	}

	resp.Say("Hello, World!")

	return nil
}
