package history

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fioncat/kubewrap/pkg/dirs"
	"github.com/icza/backscanner"
)

type manager struct {
	path    string
	records []*Record
}

func NewManager(path string, max int) (Manager, error) {
	mgr := &manager{path: path}
	err := mgr.scan(max)
	if err != nil {
		return nil, err
	}
	return mgr, nil
}

func (m *manager) scan(max int) error {
	file, err := os.Open(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
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

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		record, ok := m.parse(line)
		if !ok {
			continue
		}
		m.records = append(m.records, record)
		if len(m.records) >= max {
			break
		}
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
	}, true
}

func (m *manager) Add(name, namespace string) {
	m.records = append(m.records, &Record{
		Timestamp: time.Now().Unix(),
		Name:      name,
		Namespace: namespace,
	})
}

func (m *manager) GetLastName(current string) *string {
	for i := len(m.records) - 1; i >= 0; i-- {
		if m.records[i].Name != current {
			return &m.records[i].Name
		}
	}
	return nil
}

func (m *manager) GetLastNamespace(name, current string) *string {
	for i := len(m.records) - 1; i >= 0; i-- {
		record := m.records[i]
		if record.Namespace == "" {
			continue
		}
		if record.Name == name && record.Namespace != current {
			return &record.Namespace
		}
	}
	return nil
}

func (m *manager) DeleteByName(name string) {
	newRecords := make([]*Record, 0)
	for _, record := range m.records {
		if record.Name == name {
			continue
		}
		newRecords = append(newRecords, record)
	}
	m.records = newRecords
}

func (m *manager) DeleteAll() {
	m.records = nil
}

func (m *manager) List() []*Record {
	return m.records
}

func (m *manager) Save() error {
	err := dirs.EnsureCreate(filepath.Dir(m.path))
	if err != nil {
		return fmt.Errorf("ensure history directory: %w", err)
	}

	file, err := os.Create(m.path)
	if err != nil {
		return fmt.Errorf("create history file: %w", err)
	}
	defer file.Close()

	for _, record := range m.records {
		sb := strings.Builder{}
		sb.WriteString(fmt.Sprint(record.Timestamp))
		sb.WriteByte(' ')
		sb.WriteString(record.Name)
		if record.Namespace != "" {
			sb.WriteByte(' ')
			sb.WriteString(record.Namespace)
		}
		sb.WriteByte('\n')
		_, err = file.WriteString(sb.String())
		if err != nil {
			return fmt.Errorf("write history file: %w", err)
		}
	}

	return nil
}
