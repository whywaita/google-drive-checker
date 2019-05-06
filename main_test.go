package main

import (
	"log"
	"os"
	"testing"

	"github.com/johntdyer/slackrus"
	"github.com/sirupsen/logrus"

	"google.golang.org/api/drive/v3"

	"github.com/whywaita/google-drive-checker/api"
	"github.com/whywaita/google-drive-checker/checker"
	"github.com/whywaita/google-drive-checker/config"
	"github.com/whywaita/google-drive-checker/files"
)

var (
	con  = &config.Config{}
	srv  = &drive.Service{}
	dirs = &[]*files.Files{}
)

func TestMain(m *testing.M) {
	before()
	os.Exit(m.Run())
}

func before() {
	c, err := config.Initiallize()
	if err != nil {
		log.Fatal(err)
	}
	con = c

	client := api.GetClient(*con)

	s, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	srv = s

	ds, err := files.GetAllDirList(srv, con.ParentId)
	if err != nil {
		log.Fatalf("Unable to get all directory List: %v", err)
	}
	dirs = &ds

	logrus.AddHook(&slackrus.SlackrusHook{
		HookURL:        con.SlackHookURL,
		AcceptedLevels: slackrus.LevelThreshold(logrus.InfoLevel),
		Channel:        con.SlackChannelName,
		IconEmoji:      ":ghost:",
		Username:       "Google Drive Checker",
	})
}

func TestDuplicateDirName(t *testing.T) {
	detected := checker.DuplicateDirName(*dirs)
	if detected != nil {
		for folderName, ids := range detected {
			logrus.Infof("%v: %v\n", folderName, ids)
		}
		t.Fatal("detected duplicate directory name!")
	}
}

func TestZeroByteFile(t *testing.T) {
	code := 0
	var fs []*files.Files

	err := files.GetFileListByDirs(srv, &fs, *dirs)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range fs {
		err := checker.ZeroByteFile(f)
		if err != nil {
			// notify slack
			logrus.Warn(err)
			code = 1
		}
	}

	if code == 1 {
		t.Fatal("detected size of file is zero!")
	}

}
