package email

import (
	"crypto/tls"
	"errors"
	"net/smtp"
	"os"
	"pharmacy/models"
	"pharmacy/services/database"
	"strconv"
)

var (
	ErrorEmptyEmail   = errors.New("mail is empty")
	ErrorEmptyMessage = errors.New("message is empty")
	ErrorInvalidEmail = errors.New("invalid email")
)

type EmailService struct {
	Database *database.DatabaseService
}

func (e *EmailService) CreateOrderBody(order models.Order) {
	var basket_contents []models.BasketContentDTO

	e.Database.DB.Model(&models.BasketContent{}).
		Select("products.name ,basket_contents.count").
		Joins("join products on basket_contents.product_id = products.id").
		Where("basket_id = ?", order.BasketID).Scan(&basket_contents)

	body := "Заказ №" + strconv.Itoa(int(order.ID)) + " сформирован. Список покупок: \n"

	for _, product := range basket_contents {
		body += product.Name + ". В количестве: " + strconv.Itoa(int(product.Count)) + " шт\n"
	}

	var email []string

	e.Database.DB.Model(&models.User{}).Where("id = ?", order.UserID).Pluck("email", &email)
	e.SendMessage(email, body)
	var admin_email []string
	admin_email = append(admin_email, os.Getenv("EMAIL"))
	e.SendMessage(admin_email, body)
}

func (e *EmailService) SendMessage(email_receivers []string, message string) error {
	admin_email := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")
	host := os.Getenv("HOST")
	port := os.Getenv("EMAIL_PORT")

	auth := smtp.PlainAuth("", admin_email, password, host)
	if len(email_receivers) == 0 {
		return ErrorEmptyEmail
	}
	for _, item := range email_receivers {
		if item == "" {
			return ErrorEmptyEmail
		}
	}

	var conf = &tls.Config{ServerName: host}

	conn, connErr := tls.Dial("tcp", host+":"+port, conf)
	if connErr != nil {
		return connErr
	}

	cl, clErr := smtp.NewClient(conn, host)
	if clErr != nil {
		return clErr
	}

	auntErr := cl.Auth(auth)
	if auntErr != nil {
		return auntErr
	}

	mailErr := cl.Mail(admin_email)
	if mailErr != nil {
		return mailErr
	}

	rcptErr := cl.Rcpt(email_receivers[0])
	if rcptErr != nil {
		return rcptErr
	}

	var w, err = cl.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	err = cl.Quit()
	if err != nil {
		return err
	}

	return nil
}
