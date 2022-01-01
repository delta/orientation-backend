package webhooks

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/webhook"
	"google.golang.org/protobuf/encoding/protojson"
)

func RegisterRoutes(v *echo.Group) {
	v.POST("/webhook/events", GetEvents)
}

func GetEvents(c echo.Context) error {
	fmt.Println("Enters here")
	data, err := webhook.Receive(c.Request(), auth.NewFileBasedKeyProviderFromMap(map[string]string{"APIwFogrEH7TpB": "nAB2w0sdv313XKjz1ee3ZOVrc2CBrN1npLPEH033nwv"}))
	if err != nil {
		fmt.Println("Error")
		return err
	}
	event := livekit.WebhookEvent{}
	if err = protojson.Unmarshal(data, &event); err != nil {
		return err
	}
	fmt.Println(event.GetEvent())
	return err
}
