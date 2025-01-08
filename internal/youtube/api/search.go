package api

func (ytApi Youtube) Search(query string, single bool) ([]string, error) {
	var limit int64 = 1
	if !single {
		limit = int64(ytApi.searchLimit)
	}
	resp, err := ytApi.serv.Search.List([]string{"id", "snippet"}).Type("video").MaxResults(limit).Order("relevance").SafeSearch("none").Q(query).Do()
	if err != nil {
		return nil, err
	}

	res := make([]string, 0, limit)
	for _, item := range resp.Items {
		res = append(res, item.Id.VideoId+":"+item.Snippet.Title)
	}

	return res, nil
}
