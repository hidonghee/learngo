package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id             string
	title          string
	location       string
	career         string
	education      string
	employmentType string
	salary         string
	keywordSlice   []string
}

// 잡코리아의 go에 대한 채용 리스트
// 초기 페이지는 1임

// Scrape Indeed by a term
func Scrape(term string) { // term은 검색어를 받을 인자
	var baseURL string = "https://www.saramin.co.kr/zf_user/search?searchType=search&searchword=" + term + "&recruitPage=" // 페이지는 1 부터 시작
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)
	for i := 1; i < totalPages+1; i++ {
		go getPage(baseURL, i, c)
	}
	for i := 1; i < totalPages+1; i++ {
		extractedJobs := <-c                  // 각 페이지의 모든 구인공고(카드들)를 담은 배열을 모아둔 배열 => 1~6페이지 까지 각 구인공고 배열을 한 배열에 담음.
		jobs = append(jobs, extractedJobs...) // 두개의 배열을 합치려면 '...'의 키워드를 넣어주면 된다.
	}
	writeJobs(jobs)
	fmt.Println("Done, extraced", len(jobs))
}

// 검색 결과로 출력된 한 페이지 기준 모든 구인공고를 가져옴
func getPage(baseURL string, page int, mainC chan<- []extractedJob) { // => only send channel
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := baseURL + strconv.Itoa(page)
	fmt.Println("Requesting :", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".item_recruit") //검색 결과 페이지의 구인공고 카드의 html class
	searchCards.Each(func(i int, s *goquery.Selection) {
		go extractJob(s, c)

	})
	for i := 0; i < searchCards.Length(); i++ {
		job := <-c               // 한 일자리(카드)를 job
		jobs = append(jobs, job) // 한 일자리카드를 계속해서 담을 변수
	}
	mainC <- jobs
}

// 검색에 대한 전체 페이지 숫자를 가져옴.
func getPages(baseURL string) int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	// res.Body가 바이트로 흘러오기 떄문에 메모리 누수를 막기 위해 close해주자.
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Children().Length()
	})

	return pages
}

// 카드에서 직접 추출하기 => 한 구인리스트(카드)를 가져오는 함수
func extractJob(s *goquery.Selection, c chan<- extractedJob) {
	id, _ := s.Attr("value")

	title, _ := s.Find(".job_tit>a").Attr("title")

	// location := s.Find(".job_condition>span>a").Text()
	// fmt.Println(location)
	location := CleanString(s.Find(".job_condition").Children().Eq(0).Text())

	career := s.Find(".job_condition").Children().Eq(1).Text()

	education := s.Find(".job_condition").Children().Eq(2).Text()

	employmentType := s.Find(".job_condition").Children().Eq(3).Text()

	salary := s.Find(".job_condition").Children().Eq(4).Text()

	sectorLegnth := s.Find(".job_sector").Children().Length()
	keyword := make([]string, sectorLegnth-1)
	for i := 0; i < sectorLegnth-1; i++ {
		keyword[i] = s.Find(".job_sector").Children().Eq(i).Text()
	}
	c <- extractedJob{
		id:             id,
		title:          title,
		location:       location,
		career:         career,
		education:      education,
		employmentType: employmentType,
		salary:         salary,
		keywordSlice:   keyword}

}

// 지속적으로 에러체크를 해야하기에 에러체크함수
func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request Failed with Status", res.StatusCode)
	}

}

// CleanString cleans a string
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

// csv로 저장하는 함수 => go의 패키지를 사용한다. => https://www.convertcsv.com/csv-viewer-editor.htm 링크 이용하기
func writeJobs(jobs []extractedJob) {
	// go의 os 패키지
	file, err := os.Create("jobs.csv")
	checkErr(err)
	w := csv.NewWriter(file)
	defer w.Flush()    // 함수가 끝나는 시점에 파일에 데이터 입력
	defer file.Close() // 입력이 끝나고 파일을 닫음
	headers := []string{"LINK", "TITLE", "LOCATION", "CAREER", "EDUCATION", "EMPLOYMENT_TYPE", "SALARY", "KEYWORD"}
	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://www.saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx=" + job.id, job.title, job.location, job.career, job.education, job.employmentType, job.salary, job.keywordSlice[0]}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

/*과제 csv파일을 만드는 것도 고루틴이 가능 한번 도전해보기 => 과제*/
