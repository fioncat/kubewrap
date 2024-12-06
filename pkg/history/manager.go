package history

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/icza/backscanner"
)

type manager struct {
	path    string
	records []*Record
}

func (m *manager) scan(max int) error {
	file, err := os.Open(m.path)
	if err != nil {
		return fmt.Errorf("open history file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	scanner := backscanner.New(file, int(stat.Size()))
	for {
		line, _, err := scanner.Line()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("scan history file: %w", err)
		}
		if len(m.records) >= max {
			break
		}

		record, ok := m.parse(line)
		if !ok {
			continue
		}
		m.records = append(m.records, record)
	}

	return nil
}

func (m *manager) parse(line string) (*Record, bool) {
	fields := strings.Fields(line)

	if len(fields) != 2 && len(fields) != 3 {
		return nil, false
	}

	timestamp, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return nil, false
	}
	if timestamp == 0 {
		return nil, false
	}

	name := fields[1]
	if name == "" {
		return nil, false
	}

	var namespace string
	if len(fields) == 3 {
		namespace = fields[2]
	}

	return &Record{
		Timestamp: timestamp,
		Name:      name,
		Namespace: namespace,
	}, false
}

func (m *manager) Add(name, namespace string) {
	m.records = append(m.records, &Record{
		Timestamp: time.Now().Unix(),
		Name:      name,
		Namespace: namespace,
	})
}
