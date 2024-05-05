package sam

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const URL_BASE = "https://api.sam.gov/opportunities/v2/search"

type Sam struct {
	API_KEY string
}

type Listings struct {
	TotalRecords      int       `json:"totalRecords"`
	Limit             int       `json:"limit"`
	Offset            int       `json:"offset"`
	OpportunitiesData []Listing `json:"opportunitiesData"`
}

type Listing struct {
	NoticeId string `json:"noticeId"`
	Title    string `json:"title"`
}

func (sam *Sam) GetNListings(n int) (*[]Listing, error) {
	url := URL_BASE + "?api_key=" + sam.API_KEY + "&limit=" + strconv.Itoa(n) + "&postedFrom=01/01/2018&postedTo=05/10/2018"

	fmt.Println("Making GET Request...")
	resp, err := http.Get(url)

	if err != nil {
		return &[]Listing{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return &[]Listing{}, err
	}

	data := Listings{}
	json.Unmarshal(body, &data)
	fmt.Println(url)
	fmt.Println(data)

	return &data.OpportunitiesData, nil
}
