package api

func (s Youtube) Search(query string, limit int) (*[]string, error) {
	resp, err := s.serv.Search.List([]string{"id", "snippet"}).Type("video").MaxResults(int64(limit)).Order("relevance").SafeSearch("none").Q(query).Do()
	if err != nil {
		return nil, err
	}

	res := make([]string, 0)
	for _, item := range resp.Items {
		res = append(res, item.Id.VideoId+" "+item.Snippet.Title)
	}
	return &res, nil
}
