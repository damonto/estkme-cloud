package lpac

type DiscoveryResponse struct {
	RspServerAddress string `json:"rspServerAddress"`
}

func (c *Cmder) Discovery() ([]DiscoveryResponse, error) {
	var response []DiscoveryResponse
	if err := c.Run([]string{"profile", "discovery"}, &response, nil); err != nil {
		return response, err
	}
	return response, nil
}
