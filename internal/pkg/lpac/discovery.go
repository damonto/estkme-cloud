package lpac

type DiscoveryResponse struct {
	RspServer string `json:"rspServer"`
}

func (c *cli) Discovery() (DiscoveryResponse, error) {
	var response DiscoveryResponse
	if err := c.Run([]string{"discovery"}, &response, nil); err != nil {
		return response, err
	}
	return response, nil
}
