package scrape

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Student struct {
	NIM           string
	Name          string
	Major         string
	Semester      string
	MaximumCredit string
}

type Identity struct {
	NIM      string `json:"nim"`
	Password string `json:"password"`
}

type StudyPlan struct {
	Code       string
	Class      string
	CourseName string
	Credits    string
}

func (s *Scrape) Login(ctx context.Context, identity Identity) error {
	err := chromedp.Run(s.ctx, chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.ClearBrowserCookies().Do(ctx)
		}),
		chromedp.Navigate("https://portal.uad.ac.id/"),
		chromedp.WaitVisible("input[name='login']"),
		chromedp.WaitVisible("input[name='password']"),
		chromedp.SendKeys("input[name='login']", identity.NIM),
		chromedp.SendKeys("input[name='password']", identity.Password),
		chromedp.Click("button[type='submit']"),
		chromedp.Sleep(1 * time.Second),
	})
	if err != nil {
		return err
	}

	var errorExists bool
	err = chromedp.Run(s.ctx, chromedp.Tasks{
		chromedp.Evaluate(`!!document.querySelector('.help-block')`, &errorExists),
	})
	if err != nil {
		return err
	}

	if errorExists {
		return errors.New("login failed")
	}

	return nil
}

func (s *Scrape) GetStudentData(ctx context.Context) (*Student, error) {
	var student Student
	err := chromedp.Run(s.ctx, chromedp.Tasks{
		chromedp.Navigate("https://portal.uad.ac.id/krs/Krs"),
		chromedp.Text(`//tr[td[contains(text(),"NIM")]]/td[3]`, &student.NIM),
		chromedp.Text(`//tr[td[contains(text(),"Nama")]]/td[3]`, &student.Name),
		chromedp.Text(`//tr[td[contains(text(),"Program Studi")]]/td[3]`, &student.Major),
		chromedp.Text(`//tr[td[contains(text(),"Semester")]]/td[3]`, &student.Semester),
		chromedp.Text(`//tr[td[contains(text(),"Maksimum SKS")]]/td[3]`, &student.MaximumCredit),
	})
	if err != nil {
		return nil, err
	}

	// Clean up MaximumCredit to only get the number
	student.MaximumCredit = strings.Split(student.MaximumCredit, ",")[0]
	return &student, nil
}

func (s *Scrape) GetStudyPlans(ctx context.Context) ([]StudyPlan, error) {
	var studyPlans []StudyPlan
	var codes, classes, courseNames, credits []string

	err := chromedp.Run(s.ctx, chromedp.Tasks{
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('table.table-striped tbody tr')).map(row => 
				row.querySelector('td:nth-child(3)').textContent
			)
		`, &codes),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('table.table-striped tbody tr')).map(row => 
				row.querySelector('td:nth-child(2)').textContent
			)
		`, &classes),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('table.table-striped tbody tr')).map(row => 
				row.querySelector('td:nth-child(4)').textContent
			)
		`, &courseNames),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('table.table-striped tbody tr')).map(row => 
				row.querySelector('td:nth-child(5)').textContent
			)
		`, &credits),
	})
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(codes); i++ {
		studyPlans = append(studyPlans, StudyPlan{
			Code:       codes[i],
			Class:      classes[i],
			CourseName: courseNames[i],
			Credits:    credits[i],
		})
	}

	return studyPlans, nil
}
