package main

//https://echo.labstack.com/guide/ => go 웹 프레임워크 echo
import (
	"os"
	"strings"

	"github.com/hidonghee/learngo/scrapper"
	"github.com/labstack/echo"
)

const FILE_NAME string = "jobs.csv"

func handleHome(c echo.Context) error {
	// return c.String(http.StatusOK, "Hello, World!")
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	defer os.Remove(FILE_NAME)
	scrapper.Scrape(term)
	return c.Attachment(FILE_NAME, FILE_NAME) // 첨부파일을 리턴하는 기능 => jobs.csv파일을 job.csv파일로 전달할 것.
}
func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start((":1234")))
}
