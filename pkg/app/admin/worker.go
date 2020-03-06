package admin

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/qorpress/qorpress/internal/admin"
	"github.com/qorpress/qorpress/internal/exchange"
	"github.com/qorpress/qorpress/internal/exchange/backends/csv"
	"github.com/qorpress/qorpress/internal/i18n/exchange_actions"
	"github.com/qorpress/qorpress/internal/media/oss"
	"github.com/qorpress/qorpress/internal/qor"
	"github.com/qorpress/qorpress/internal/worker"

	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress/pkg/config/i18n"
	"github.com/qorpress/qorpress/pkg/models/posts"
)

// SetupWorker setup worker
func SetupWorker(Admin *admin.Admin) {
	Worker := worker.New()

	type sendNewsletterArgument struct {
		Subject      string
		Content      string `sql:"size:65532"`
		SendPassword string
		worker.Schedule
	}

	Worker.RegisterJob(&worker.Job{
		Name: "Send Newsletter",
		Handler: func(argument interface{}, qorJob worker.QorJobInterface) error {
			qorJob.AddLog("Started sending newsletters...")
			qorJob.AddLog(fmt.Sprintf("Argument: %+v", argument.(*sendNewsletterArgument)))
			for i := 1; i <= 100; i++ {
				time.Sleep(100 * time.Millisecond)
				qorJob.AddLog(fmt.Sprintf("Sending newsletter %v...", i))
				qorJob.SetProgress(uint(i))
			}
			qorJob.AddLog("Finished send newsletters")
			return nil
		},
		Resource: Admin.NewResource(&sendNewsletterArgument{}),
	})

	type importPostArgument struct {
		File oss.OSS
	}

	Worker.RegisterJob(&worker.Job{
		Name:  "Import Posts",
		Group: "Posts Management",
		Handler: func(arg interface{}, qorJob worker.QorJobInterface) error {
			argument := arg.(*importPostArgument)

			context := &qor.Context{DB: db.DB}

			var errorCount uint

			if err := PostExchange.Import(
				csv.New(filepath.Join("public", argument.File.URL())),
				context,
				func(progress exchange.Progress) error {
					var cells = []worker.TableCell{
						{Value: fmt.Sprint(progress.Current)},
					}

					var hasError bool
					for _, cell := range progress.Cells {
						var tableCell = worker.TableCell{
							Value: fmt.Sprint(cell.Value),
						}

						if cell.Error != nil {
							hasError = true
							errorCount++
							tableCell.Error = cell.Error.Error()
						}

						cells = append(cells, tableCell)
					}

					if hasError {
						if errorCount == 1 {
							var headerCells = []worker.TableCell{
								{Value: "Line No."},
							}
							for _, cell := range progress.Cells {
								headerCells = append(headerCells, worker.TableCell{
									Value: cell.Header,
								})
							}
							qorJob.AddResultsRow(headerCells...)
						}

						qorJob.AddResultsRow(cells...)
					}

					qorJob.SetProgress(uint(float32(progress.Current) / float32(progress.Total) * 100))
					qorJob.AddLog(fmt.Sprintf("%d/%d Importing post %v", progress.Current, progress.Total, progress.Value.(*posts.Post).Code))
					return nil
				},
			); err != nil {
				qorJob.AddLog(err.Error())
			}

			return nil
		},
		Resource: Admin.NewResource(&importPostArgument{}),
	})

	Worker.RegisterJob(&worker.Job{
		Name:  "Export Posts",
		Group: "Posts Management",
		Handler: func(arg interface{}, qorJob worker.QorJobInterface) error {
			qorJob.AddLog("Exporting posts...")

			context := &qor.Context{DB: db.DB}
			fileName := fmt.Sprintf("/downloads/posts.%v.csv", time.Now().UnixNano())
			if err := PostExchange.Export(
				csv.New(filepath.Join("public", fileName)),
				context,
				func(progress exchange.Progress) error {
					qorJob.AddLog(fmt.Sprintf("%v/%v Exporting post %v", progress.Current, progress.Total, progress.Value.(*posts.Post).Code))
					return nil
				},
			); err != nil {
				qorJob.AddLog(err.Error())
			}

			qorJob.SetProgressText(fmt.Sprintf("<a href='%v'>Download exported posts</a>", fileName))
			return nil
		},
	})

	exchange_actions.RegisterExchangeJobs(i18n.I18n, Worker)
	Admin.AddResource(Worker, &admin.Config{Menu: []string{"Site Management"}, Priority: 3})
}
