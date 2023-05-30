package file_storage

import (
	"bufio"
	"encoding/json"
	"github.com/volnistii11/URL-shortener/internal/app/config"
	"github.com/volnistii11/URL-shortener/internal/app/storage"
	"os"
)

type Event struct {
	ID          uint   `json:"uuid,string"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteEvent(event *Event) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file *os.File
	// добавляем reader в Consumer
	reader *bufio.Reader
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый Reader
		reader: bufio.NewReader(file),
	}, nil
}

func (c *Consumer) ReadEvent() (*Event, error) {
	// читаем данные до символа переноса строки
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// преобразуем данные из JSON-представления в структуру
	event := Event{}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

type Restorer interface {
	RestoreDataFromJSONFileToStructure() storage.Repository
}

func NewSetter(repository storage.Repository, cfg config.Flags) Restorer {
	return &restoredURL{
		repo:  repository,
		flags: cfg,
	}
}

type restoredURL struct {
	repo  storage.Repository
	flags config.Flags
}

func (r *restoredURL) RestoreDataFromJSONFileToStructure() storage.Repository {
	Consumer, _ := NewConsumer(r.flags.GetFileStoragePath())
	defer Consumer.Close()
	for {
		readEvent, _ := Consumer.ReadEvent()
		if readEvent == nil {
			break
		}
		r.repo.SetRestoreData(readEvent.ShortURL, readEvent.OriginalURL)
	}
	return r.repo
}
