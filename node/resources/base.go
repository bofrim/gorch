package resources

type ResourceBase struct {
	Name  string
	Count int64
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
