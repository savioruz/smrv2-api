package scrape

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Faculty struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

type StudyProgram struct {
	Value   string `json:"value"`
	Name    string `json:"name"`
	Faculty string `json:"faculty"`
}

type Schedule struct {
	Hari     string `json:"hari"`
	Kode     string `json:"kode"`
	Matkul   string `json:"matkul"`
	Kelas    string `json:"kelas"`
	Sks      string `json:"sks"`
	Jam      string `json:"jam"`
	Semester string `json:"semester"`
	Dosen    string `json:"dosen"`
	Ruang    string `json:"ruang"`
}

func (s *Scrape) GetStudyPrograms(ctx context.Context) (map[string][]StudyProgram, error) {
	var faculty []Faculty
	var studyPrograms = make(map[string][]StudyProgram)

	err := chromedp.Run(s.ctx, chromedp.Tasks{
		chromedp.Navigate("https://simeru.uad.ac.id/?mod=auth&sub=auth"),

		chromedp.SendKeys(`input[name="user"]`, "mhs"),
		chromedp.SendKeys(`input[name="pass"]`, "mhs"),
		chromedp.Click(`input[type="submit"]`),
		chromedp.Sleep(2 * time.Second),

		chromedp.Navigate("https://simeru.uad.ac.id/?mod=laporan_baru&sub=jadwal_prodi&do=daftar"),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('select[name="fakultas"] option')).map(option => ({ value: option.value, name: option.text }))`, &faculty),

		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, f := range faculty {
				if f.Value == "" {
					continue
				}

				err := chromedp.Run(ctx, chromedp.SetValue(`select[name="fakultas"]`, f.Value))
				if err != nil {
					return fmt.Errorf("error setting faculty value: %w", err)
				}

				err = chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`popUpFak(document.form_cari.prodi, %q)`, f.Value), nil))
				if err != nil {
					return fmt.Errorf("error triggering popUpFak: %w", err)
				}

				var programs []StudyProgram
				err = chromedp.Run(ctx, chromedp.Evaluate(
					`Array.from(document.querySelectorAll('select[name="prodi"] option')).filter(option => option.value !== "").map(option => ({ value: option.value, name: option.text }))`,
					&programs,
				))
				if err != nil {
					return fmt.Errorf("error getting study programs: %w", err)
				}

				// Add faculty name to each program
				for i := range programs {
					programs[i].Faculty = strings.ToLower(f.Name)
				}

				studyPrograms[f.Value] = programs
			}
			return nil
		}),
	})

	if err != nil {
		return nil, err
	}

	return studyPrograms, nil
}

func (s *Scrape) GetSchedule(ctx context.Context, facultyID, programID string) ([]Schedule, error) {
	var schedules []Schedule

	err := chromedp.Run(s.ctx,
		chromedp.Navigate("https://simeru.uad.ac.id/?mod=auth&sub=auth"),
		chromedp.SendKeys(`input[name="user"]`, "mhs"),
		chromedp.SendKeys(`input[name="pass"]`, "mhs"),
		chromedp.Click(`input[type="submit"]`),
		chromedp.Sleep(2*time.Second),

		chromedp.Navigate("https://simeru.uad.ac.id/?mod=laporan_baru&sub=jadwal_prodi&do=daftar"),
		chromedp.SetValue(`select[name=fakultas]`, facultyID, chromedp.NodeVisible),
		chromedp.SetValue(`select[name=prodi]`, programID, chromedp.NodeVisible),
		chromedp.Click(`input[name=submit]`, chromedp.NodeVisible),
		chromedp.WaitVisible(`table.table-list`),

		chromedp.Evaluate(`
		[...document.querySelectorAll("table.table-list tr")].slice(1).map(tr => {
			const cells = tr.children;
			return {
				Hari: cells[0] ? cells[0].textContent.trim() : "",
				Kode: cells[1] ? cells[1].textContent.trim() : "",
				Matkul: cells[2] ? cells[2].textContent.trim() : "",
				Kelas: cells[3] ? cells[3].textContent.trim() : "",
				Sks: cells[4] ? cells[4].textContent.trim() : "",
				Jam: cells[5] ? cells[5].textContent.trim() : "",
				Semester: cells[6] ? cells[6].textContent.trim() : "",
				Dosen: cells[7] ? cells[7].textContent.trim() : "",
				Ruang: cells[8] ? cells[8].textContent.trim() : ""
			}
		})`, &schedules))

	if err != nil {
		return nil, fmt.Errorf("error getting schedule: %w", err)
	}

	// Fill in missing days with the last known day
	var lastDay string
	for i := range schedules {
		if schedules[i].Hari != "" {
			lastDay = schedules[i].Hari
		} else {
			schedules[i].Hari = lastDay
		}
	}

	return schedules, nil
}
