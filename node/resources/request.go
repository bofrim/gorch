package resources

import "encoding/json"

type ResourceRequest struct {
	ResourceBase
}

func (r *ResourceRequest) UnmarshalJSON(data []byte) error {

	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	b.Price, _ = v[0].(string)
	b.Size, _ = v[1].(string)
	b.NumOrders = int(v[2].(float64))

	return nil
}
