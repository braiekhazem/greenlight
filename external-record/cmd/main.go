package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	_ "github.com/lib/pq"

	"os"
	"path/filepath"
	"runtime"
	"ta_video_progress_svc/pkg/dropbox"
	"ta_video_progress_svc/pkg/logger"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var build = "develop"
var (
	recordingUploadSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "recording_upload_success_total",
			Help: "Total number of successful recording uploads",
		})
	recordingUploadFailure = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "recording_upload_failure_total",
			Help: "Total number of failed recording uploads",
		})
	recordingFileSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "recording_file_size_bytes",
			Help: "Size of the recording file before uploading",
		})
	uploadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "upload_duration_seconds",
			Help:    "Time taken to upload recording file to Dropbox",
			Buckets: prometheus.DefBuckets,
		})
	apiCallSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_call_success_total",
		Help: "Total number of successful API calls after uploading",
	})
	apiCallFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_call_failure_total",
		Help: "Total number of failed API calls after uploading",
	})

	// New counters for successful and failed deletions
	deletionSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "deletion_success_total",
		Help: "Total number of successful deletions after API call",
	})
	deletionFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "deletion_failure_total",
		Help: "Total number of failed deletions after API call",
	})
)

func main() {
	log, err := logger.New("BBB_RECORDING")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}

}

func run(log *zap.SugaredLogger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		CRON struct {
			Schedule string `conf:"default:*/4 * * * *"`
		}
		Recording struct {
			Path         string `conf:"default:/var/bigbluebutton/published/video"`
			ShouldDelete bool   `conf:"default:true"`
		}
		DropBox struct {
			RefreshToken string `conf:"default:xx"`
			AppKey       string `conf:"default:xx"`
			AppSecret    string `conf:"default:xx"`
		}
		Postgres struct {
			Host     string `conf:"default:postgres"`
			Port     int    `conf:"default:5432"`
			Username string `conf:"default:postgres"`
			Password string `conf:"default:secret,mask"`
			Database string `conf:"default:greenlight"`
			SSLMode  string `conf:"default:disable"`
		}
		MetricServer struct {
			Port string `conf:"default:2113"`
		}
	}{
		Version: conf.Version{
			Build: build,
		},
	}

	const prefix = "BBB"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// -------------------------------------------------------------------------
	// connect to postgres
	log.Infow("startup", "status", "initializing postgres support")
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port,
		cfg.Postgres.Username, cfg.Postgres.Password,
		cfg.Postgres.Database, cfg.Postgres.SSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("pinging postgres: %w", err)
	}
	defer db.Close()
	// -------------------------------------------------------------------------

	// -------------------------------------------------------------------------
	// CRON JOB SCHEDULING
	log.Infow("start running cron job", "schedule", cfg.CRON.Schedule)
	dropBoxClient := dropbox.NewDropBox(cfg.DropBox.AppKey, cfg.DropBox.AppSecret, cfg.DropBox.RefreshToken)
	c := cron.New()
	_, err = c.AddFunc(cfg.CRON.Schedule, func() {

		meetingIDs := make([]string, 0)
		log.Info("Searching for new recordings in directory", "path", cfg.Recording.Path)

		files, err := ioutil.ReadDir(cfg.Recording.Path)
		if err != nil {
			log.Error(err, "Error reading directory", "path", cfg.Recording.Path)
			return
		}

		for _, file := range files {
			if file.IsDir() {
				meetingIDs = append(meetingIDs, file.Name())
			}
		}
		log.Infow("Finished searching for new recordings", "meetingIDs", meetingIDs)

		at, err := dropBoxClient.GetAccessToken()
		if err != nil {
			log.Error(err, "Error getting Dropbox access token")
			return
		}

		for _, meetID := range meetingIDs {
			log.Infow("Starting upload process for meeting recordings", "meetingID", meetID)

			// Define the meeting directory
			meetingDir := filepath.Join(cfg.Recording.Path, meetID)
			files, err := ioutil.ReadDir(meetingDir)
			if err != nil {
				log.Error(err, "Error reading meeting directory", "directory", meetingDir)
				continue
			}

			// Iterate over files in the meeting directory
			for _, file := range files {
				startUpload := time.Now()
				if !file.IsDir() && len(file.Name()) > 6 && file.Name()[:6] == "video-" && isValidVideoExtension(file.Name()) {
					log.Infow("Found video file to upload", "file", file.Name(), "meetingID", meetID)

					outputFile := fmt.Sprintf("%s/%s", meetingDir, file.Name())

					log.Infow("Uploading file to Dropbox", "file", outputFile, "meetingID", meetID)
					if recordingUrl, err := dropBoxClient.UploadFile(at, outputFile, meetID); err != nil || recordingUrl == "" {
						log.Error(err, "Error uploading file to Dropbox", "file", outputFile)
						recordingUploadFailure.Inc()
					} else {
						log.Infow("File successfully uploaded to Dropbox", "file", outputFile)
						recordingUploadSuccess.Inc()
						fileSize := file.Size()
						recordingFileSize.Set(float64(fileSize))
						uploadDuration.Observe(time.Since(startUpload).Seconds())
						// send api request to with meetingID and dropbox link
						log.Infow("Sending recording link to API", "meetingID", meetID, "recordingUrl", recordingUrl)
						err = persistLinkToDb(db, meetID, recordingUrl)
						if err != nil {
							log.Error(err, "Error persisting recording link to DB", "file", outputFile)
							apiCallFailure.Inc()
							continue
						}
						apiCallSuccess.Inc()
						log.Infow("Successfully persisti recording link to db", "meetID", meetID, "recordingUrl", recordingUrl)

					}
				}
			}
			log.Infow("Finished upload process for meeting recordings", "meetingID", meetID)
		}
		log.Info("Finished uploading all recordings , waiting for next cron job")
	})
	if err != nil {
		log.Error(err, "Error adding cron job")
		return err
	}

	c.Start()
	defer c.Stop()
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "OK")
	})

	err = http.ListenAndServe(fmt.Sprintf(":%s", cfg.MetricServer.Port), nil)
	if err != nil {
		fmt.Println("error starting prometheus server ", err)
	}
	return nil

}

func isValidVideoExtension(filename string) bool {
	validExtensions := []string{".mp4", ".m4v", ".webm", ".mov", ".avi"}

	ext := filepath.Ext(filename)
	// Check if the extension is in the list of valid extensions
	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

func persistLinkToDb(db *sql.DB, meetingID, recordingUrl string) error {
	query := `
		UPDATE formats f
		SET url = $1
		FROM recordings AS r
		WHERE f.recording_id = r.id 
		AND r.record_id= $2;	
	
	`
	_, err := db.Exec(query, recordingUrl, meetingID)
	return err
}
