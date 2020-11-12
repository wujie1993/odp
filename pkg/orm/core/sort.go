package core

type SortByCreateTime ApiObjectList

func (s SortByCreateTime) Len() int {
	return len(s)
}

func (s SortByCreateTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortByCreateTime) Less(i, j int) bool {
	return s[i].GetMetadata().CreateTime.Before(s[j].GetMetadata().CreateTime)
}
