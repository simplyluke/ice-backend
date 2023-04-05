package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type emergency_plan struct {
	Destination      string `json:"destination"`
	GroupMembers     string `json:"group_members"`
	EmergencyContact string `json:"emergency_contact"`
	Trailhead        string `json:"trailhead"`
	Car              string `json:"car"`
	Clothing         string `json:"clothing"`
	Equipment        string `json:"equipment"`
	PreparedNight    bool   `json:"prepared_night"`
	ExpectedReturn   string `json:"expected_return"`
	EmergencyTime    string `json:"emergency_time"`
	RecipientEmail   string `json:"recipient_email"`
}

func main() {
	godotenv.Load()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://simplyluke.com"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	authorized := router.Group("/api", gin.BasicAuth(gin.Accounts{
		os.Getenv("AUTHED_USER"): os.Getenv("AUTHED_PASSWORD"),
	}))
	authorized.POST("/authorize", authenticate)
	authorized.POST("/emergency_plan", postEmergencyPlan)

	router.Run()
}

func postEmergencyPlan(c *gin.Context) {
	var plan emergency_plan
	if err := c.BindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, sendgridError := sendEmail(plan)
	if sendgridError != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sendgridError.Error()})
	}
	c.JSON(response.StatusCode, gin.H{"sendgrid_response": response.Body})
}

func sendEmail(plan emergency_plan) (*rest.Response, error) {
	senderEmail := os.Getenv("SENDER_EMAIL")
	from := mail.NewEmail("Luke Wright", senderEmail)
	subject := "Luke Wright has just shared an emergency plan with you"
	to := mail.NewEmail(plan.RecipientEmail, plan.RecipientEmail)
	var preparedNight string
	if plan.PreparedNight {
		preparedNight = "Yes"
	} else {
		preparedNight = "No"
	}
	htmlContent := "<p>Luke Wright has just shared an emergency plan with you.</p><p>Destination: " + plan.Destination + "</p><p>Group Members: " + plan.GroupMembers + "</p><p>Emergency Contact: " + plan.EmergencyContact + "</p><p>Trailhead: " + plan.Trailhead + "</p><p>Car: " + plan.Car + "</p><p>Clothing: " + plan.Clothing + "</p><p>Equipment: " + plan.Equipment + "</p><p>Prepared for a night out: " + preparedNight + "</p><p>Expected Return: " + plan.ExpectedReturn + "</p><p>Emergency Time: " + plan.EmergencyTime + "</p>"
	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	return client.Send(message)
}

func authenticate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Authorized"})
}
