package resources

type ResourceBase struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func NewResourceBaseMap(m map[string]int64) map[string]*ResourceBase {
	out := map[string]*ResourceBase{}
	for name, count := range m {
		out[name] = &ResourceBase{
			Name:  name,
			Count: count,
		}
	}
	return out
}
