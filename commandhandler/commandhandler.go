package commandhandler

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"whatsapp_multi_session/config"
	"whatsapp_multi_session/primitive"
	"whatsapp_multi_session/proxy"
	"whatsapp_multi_session/utils"

	"github.com/mdp/qrterminal/v3"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var (
	ClientConcurrent sync.Map
)

type CommandHandler struct {
	Container    *sqlstore.Container
	ProxyManager *proxy.Manager
}

func NewCommandHandler(container *sqlstore.Container, proxyManager *proxy.Manager) CommandHandler {
	return CommandHandler{
		Container:    container,
		ProxyManager: proxyManager,
	}
}

// StoreClientConcurrent Store a client in sync.Map
func StoreClientConcurrent(clientID string, client *whatsmeow.Client) {
	ClientConcurrent.Store(clientID, client)
}

// LoadClientConcurrent Load a client from sync.Map
func LoadClientConcurrent(clientID string) (*whatsmeow.Client, bool) {
	value, ok := ClientConcurrent.Load(clientID)
	if value == nil && !ok {
		return nil, false
	}
	// Type assertion to convert interface{} back to *whatsmeow.Client
	client, ok := value.(*whatsmeow.Client)
	if client == nil && !ok {
		return nil, false
	}
	return client, ok
}

// DeleteClientConcurrent Delete a client from sync.Map
func DeleteClientConcurrent(clientID string) {
	ClientConcurrent.Delete(clientID)
}

func (ch CommandHandler) HandleCheckUser(sender types.JID, recipientPhones []string) (response []types.IsOnWhatsAppResponse) {
	if len(recipientPhones) < 1 {
		log.Errorf("usage: checkuser <phone numbers...>")
		return nil
	}

	client, ok := LoadClientConcurrent(sender.User)
	if !ok || client == nil {
		log.Errorf("could not find client for sender %v", sender.User)
		return nil
	}

	resp, err := client.IsOnWhatsApp(recipientPhones)
	if err != nil {
		log.Errorf("failed to check if users are on WhatsApp: %v", err)
		return nil
	}

	if len(resp) > 0 {
		for _, item := range resp {
			logMessage := fmt.Sprintf("%s: on WhatsApp: %t, JID: %s", item.Query, item.IsIn, item.JID)

			if item.VerifiedName != nil {
				logMessage += fmt.Sprintf(", business name: %s", item.VerifiedName.Details.GetVerifiedName())
			}
			log.Printf(logMessage)
			response = append(response, item)
		}
	}

	return response
}

func (ch CommandHandler) HandleSendPresence(sender types.JID) (err error) {
	client, ok := LoadClientConcurrent(sender.User)
	if !ok || client == nil {
		log.Errorf("could not find client for sender %v", sender.User)
		return nil
	}

	err = client.SendPresence(types.PresenceAvailable)
	if err != nil {
		log.Errorf("error sending presence: %v", err)
		return
	}
	client.AddEventHandler(ch.eventHandler)
	return nil
}

func (ch CommandHandler) HandleGetAllDevices(ctx context.Context) (response []primitive.Devices) {
	container, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("failed to check if users are on WhatsApp: %v", err)
		return nil
	}

	if len(container) > 0 {
		for _, item := range container {
			var isLoggedIn bool
			client, ok := LoadClientConcurrent(item.ID.User)
			if ok {
				isLoggedIn = client.IsLoggedIn()
			} else {
				isLoggedIn = false
			}
			newItem := primitive.Devices{
				PushName:   item.PushName,
				Platform:   item.Platform,
				User:       item.ID.User,
				Server:     item.ID.Server,
				IsLoggedIn: isLoggedIn,
			}
			response = append(response, newItem)
		}
	} else {
		emptyResp := make([]primitive.Devices, 0)
		response = emptyResp
	}
	return response
}

// HandleDeviceProxies returns all devices along with the proxy URL they are assigned.
func (ch CommandHandler) HandleDeviceProxies(ctx context.Context) ([]primitive.DevicesWithProxy, error) {
	// fetch devices from store
	devices, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var result []primitive.DevicesWithProxy
	for _, item := range devices {
		// determine login status (reuse existing logic)
		client, ok := LoadClientConcurrent(item.ID.User)
		isLogged := ok && client.IsLoggedIn()

		// get or assign proxy for this user
		p, exists := ch.ProxyManager.GetUser(item.ID.User)
		if !exists {
			p, err = ch.ProxyManager.AddUser(item.ID.User)
			if err != nil {
				log.Printf("ProxyManager error for %s: %v", item.ID.User, err)
			}
		}

		// construct URL
		var proxyURL *url.URL
		if p.BaseURL != "" {
			proxyURL = &url.URL{
				Scheme: "socks5",
				User:   url.UserPassword(p.Username, p.Password),
				Host:   fmt.Sprintf("%s:%s", p.BaseURL, p.Port),
			}
		}

		result = append(result, primitive.DevicesWithProxy{
			PushName:   item.PushName,
			Platform:   item.Platform,
			User:       item.ID.User,
			Server:     item.ID.Server,
			ProxyURL:   proxyURL,
			IsLoggedIn: isLogged,
		})
	}

	return result, nil
}

func (ch CommandHandler) HandleGetSingleDevices(ctx context.Context, jid types.JID) (response primitive.Devices) {
	container, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("failed to check if users are on WhatsApp: %v", err)
		return primitive.Devices{}
	}

	if len(container) > 0 {
		for _, item := range container {

			jidUser := strings.TrimSpace(jid.User)
			itemIdUser := strings.TrimSpace(item.ID.User)

			if jidUser == itemIdUser {
				var isLoggedIn bool
				client, ok := LoadClientConcurrent(item.ID.User)
				if ok {
					isLoggedIn = client.IsLoggedIn()
				} else {
					isLoggedIn = false
				}

				resp := primitive.Devices{
					PushName:   item.PushName,
					Platform:   item.Platform,
					User:       item.ID.User,
					Server:     item.ID.Server,
					IsLoggedIn: isLoggedIn,
				}
				return resp
			}
		}
	}
	return response
}

func (ch CommandHandler) HandleCheckUserSingle(sender types.JID, recipient string) (response types.IsOnWhatsAppResponse, err error) {
	requestSingleIsOnWhatsapp := []string{recipient}

	client, ok := LoadClientConcurrent(sender.User)
	if !ok || client == nil {
		err = errors.New(fmt.Sprintf("Could not find client for sender %v", sender.User))
		log.Errorf("Could not find client for sender %v", sender.User)
		return
	}

	resp, err := client.IsOnWhatsApp(requestSingleIsOnWhatsapp)
	if err != nil {
		log.Errorf("failed to check if users are on whatsapp: %v", err)
		return types.IsOnWhatsAppResponse{}, err
	}

	if len(resp) > 0 {
		for _, item := range resp {
			logMessage := fmt.Sprintf("%s: on WhatsApp: %t, JID: %s", item.Query, item.IsIn, item.JID)

			if item.VerifiedName != nil {
				logMessage += fmt.Sprintf(", business name: %s", item.VerifiedName.Details.GetVerifiedName())
			}
			log.Printf(logMessage)
			response = item
		}
	}

	return response, nil
}

func (ch CommandHandler) HandleSendTextMessage(sender types.JID, textMsg string, jid string) (messageID string, err error) {
	recipient, ok := utils.ParseJID(jid)
	if !ok {
		return
	}

	client, ok := LoadClientConcurrent(sender.User)
	if !ok || client == nil {
		err = errors.New(fmt.Sprintf("Could not find client for sender %v", sender.User))
		log.Errorf("Could not find client for sender %v", sender.User)
		return
	}

	err = client.SendPresence(types.PresenceAvailable)
	if err != nil {
		log.Errorf("Error sending presence: %v", err)
		return
	}

	msg := &waE2E.Message{
		Conversation: proto.String(textMsg),
	}

	//set message id from std lib whatsmeo
	messageID = client.GenerateMessageID()

	resp, err := client.SendMessage(context.Background(), recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		log.Errorf("Error sending message: %v", err)
		return
	}

	err = client.MarkRead([]types.MessageID{resp.ID}, time.Now(), recipient, sender)
	if err != nil {
		log.Errorf("Error sending MarkRead: %v", err)
		return
	}

	client.AddEventHandler(ch.eventHandler)

	// config to delete after send
	if config.Conf.DeleteAfterSend.Enable {
		// send revoke or delete message if the flag is enabled
		_, errRevoke := client.SendMessage(context.Background(), recipient, client.BuildRevoke(recipient, types.EmptyJID, messageID))
		if errRevoke != nil {
			log.Errorf("Error sending MarkRead: %v", err)
			return
		}
	}

	return resp.ID, nil
}

func (ch CommandHandler) HandleSendTextMessageBulk(sender types.JID, textMsg string, jids []string) {
	errCh := make(chan error, len(jids)) // Buffered channel for errors
	var wg sync.WaitGroup

	// modern random API (Go 1.22+)
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()<<1)))

	for _, jid := range jids {
		// Increment the WaitGroup counter for each goroutine
		wg.Add(1)

		go func(jid string) {
			defer wg.Done() // Ensure the WaitGroup counter is decremented

			// add random delay between 100ms–2000ms before sending
			delay := time.Duration(rng.IntN(1901)+100) * time.Millisecond
			time.Sleep(delay)

			client, ok := LoadClientConcurrent(sender.User)
			if !ok || client == nil {
				errCh <- fmt.Errorf("could not find client for sender %v", sender.User)
				return
			}

			if err := client.SendPresence(types.PresenceAvailable); err != nil {
				errCh <- fmt.Errorf("error sending presence: %v", err)
				return
			}

			recipient, ok := utils.ParseJID(jid)
			if !ok {
				return
			}

			msg := &waE2E.Message{
				Conversation: proto.String(textMsg),
			}

			messageID := client.GenerateMessageID()
			resp, err := client.SendMessage(context.Background(), recipient, msg, whatsmeow.SendRequestExtra{ID: messageID})
			if err != nil {
				errCh <- fmt.Errorf("error sending message: %v", err)
				return
			}

			if err := client.MarkRead([]types.MessageID{resp.ID}, time.Now(), recipient, sender); err != nil {
				errCh <- fmt.Errorf("error sending MarkRead: %v", err)
				return
			}

			if config.Conf.DeleteAfterSend.Enable {
				if _, err := client.SendMessage(context.Background(), recipient, client.BuildRevoke(recipient, types.EmptyJID, messageID)); err != nil {
					errCh <- fmt.Errorf("error sending revoke message: %v", err)
				}
			}

		}(jid)
	}

	// Close the error channel after all goroutines are done
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Process errors from the channel as they are received
	for err := range errCh {
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
	log.Println("Bulk message sending completed")
}

func (ch CommandHandler) HandleGetSingleQR(senderJidTypes types.JID) (string, error) {
	qrCtx := context.WithoutCancel(context.Background())
	device, err := ch.Container.GetFirstDevice(qrCtx)
	if err != nil {
		return "", err
	}

	// Create a client for each device
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.AddEventHandler(ch.eventHandler)

	if config.Conf.Proxy.Enable && senderJidTypes.User != "" {
		proxyData := ch.EnabledProxy(senderJidTypes.User)
		if proxyData != nil {
			client.SetProxy(http.ProxyURL(proxyData))
		}
	}

	// Connect the client synchronously
	if client.Store.ID == nil {
		qrChan, errGetQRChannel := client.GetQRChannel(qrCtx)
		if errGetQRChannel != nil {
			return "", errGetQRChannel
		}
		err = client.Connect()
		if err != nil {
			return "", err
		}

		log.Println("Waiting for QR code or login event...")
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// Add the client to the map
				StoreClientConcurrent(senderJidTypes.User, client)
				image, errGenerateCode := utils.GenerateQRCode(evt.Code)
				if errGenerateCode != nil {
					// Log the error for debugging
					log.Println("Error generating QR code:", errGenerateCode)
					return "", errGenerateCode
				}
				return string(image), nil
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			return "", err
		}
		// Add the client to the map
		StoreClientConcurrent(device.ID.User, client)
		return "", nil
	}

	return "", nil
}

func (ch CommandHandler) HandleGetSpecificQR(jid types.JID) (string, error) {
	qrCtx := context.WithoutCancel(context.Background())
	devices, err := ch.Container.GetAllDevices(qrCtx)
	if err != nil {
		return "", err
	}

	var device *store.Device
	if len(devices) > 0 {
		for _, val := range devices {
			jidUser := strings.TrimSpace(jid.User)
			valIdUser := strings.TrimSpace(val.ID.User)

			if jidUser == valIdUser {
				device = val
			}
		}
	}

	if device == nil || device.ID.User == "" {
		device = ch.Container.NewDevice()
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.AddEventHandler(ch.eventHandler)

	if config.Conf.Proxy.Enable {
		proxyData := ch.EnabledProxy(jid.User)
		if proxyData != nil {
			client.SetProxy(http.ProxyURL(proxyData))
		}
	}

	// Connect the client synchronously
	if client.Store.ID == nil {
		qrChan, errGetQr := client.GetQRChannel(qrCtx)
		if errGetQr != nil {
			return "", errGetQr
		}
		err = client.Connect()
		if err != nil {
			return "", err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// Add the client to the map
				StoreClientConcurrent(jid.User, client)
				image, errGenerateCode := utils.GenerateQRCode(evt.Code)
				if errGenerateCode != nil {
					// Log the error for debugging
					return "", errGenerateCode
				}
				return string(image), nil
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			return "", err
		}
		// Add the client to the map
		StoreClientConcurrent(jid.User, client)
		return "", nil
	}
	return "", nil
}

func (ch CommandHandler) HandleGetSingleQRResponseCode(senderJidTypes types.JID) (string, error) {
	qrCtx := context.WithoutCancel(context.Background())
	device, err := ch.Container.GetFirstDevice(qrCtx)
	if err != nil {
		return "", err
	}

	// Create a client for each device
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.AddEventHandler(ch.eventHandler)

	if config.Conf.Proxy.Enable && senderJidTypes.User != "" {
		proxyData := ch.EnabledProxy(senderJidTypes.User)
		if proxyData != nil {
			client.SetProxy(http.ProxyURL(proxyData))
		}
	}

	// Connect the client synchronously
	if client.Store.ID == nil {
		qrChan, errQRChannel := client.GetQRChannel(qrCtx)
		if errQRChannel != nil {
			return "", errQRChannel
		}
		err = client.Connect()
		if err != nil {
			return "", err
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// Add the client to the map
				StoreClientConcurrent(senderJidTypes.User, client)
				return evt.Code, nil
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			return "", err
		}
		// Add the client to the map
		StoreClientConcurrent(device.ID.User, client)
		return "", nil
	}

	return "", nil
}

func (ch CommandHandler) HandleGetPairCode(ctx context.Context, senderJidTypes types.JID) (string, error) {
	devices, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		return "", err
	}

	var device *store.Device
	if len(devices) > 0 {
		for _, val := range devices {
			jidUser := strings.TrimSpace(senderJidTypes.User)
			valIdUser := strings.TrimSpace(val.ID.User)

			if jidUser == valIdUser {
				device = val
			}
		}
	}

	if device == nil || device.ID.User == "" {
		device = ch.Container.NewDevice()
	}

	// Create a client for each device
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.AddEventHandler(ch.eventHandler)

	if config.Conf.Proxy.Enable {
		proxyData := ch.EnabledProxy(senderJidTypes.User)
		if proxyData != nil {
			client.SetProxy(http.ProxyURL(proxyData))
		}
	}

	// cannot use store.DeviceProps.Os here, since "only common browsers/OSes are allowed",
	// see https://pkg.go.dev/go.mau.fi/whatsmeow#Client.PairPhone
	// Firefox on Linux chosen arbitrarily
	clientDisplayName := "Firefox (Linux)"
	clientType := whatsmeow.PairClientFirefox

	// Connect the client synchronously
	if client.Store == nil || client.Store.ID == nil {
		err = client.Connect()
		if err != nil {
			return "", err
		}

		pairCode, errPairPhone := client.PairPhone(ctx, senderJidTypes.User, true, clientType, clientDisplayName)
		if errPairPhone != nil {
			return "", errPairPhone
		}

		// Add the client to the map
		StoreClientConcurrent(senderJidTypes.User, client)
		return pairCode, nil

	} else {
		// Add the client to the map
		clientLocal, ok := LoadClientConcurrent(senderJidTypes.User)
		if !ok || clientLocal == nil {
			err = client.Connect()
			if err != nil {
				return "", err
			}
			StoreClientConcurrent(senderJidTypes.User, client)
		}
		return "", nil
	}
}

func (ch CommandHandler) HandleConnectSingleDevice(ctx context.Context, senderJidTypes types.JID) error {
	devices, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("ch.Container.GetDevice, got err : %v", err)
		err = errors.New("you are not logged in in this server before or just do the pair phone or scan qr")
		return err
	}

	var device *store.Device
	for _, singleDevice := range devices {
		if strings.EqualFold(singleDevice.ID.User, senderJidTypes.User) {
			device = singleDevice
		}
	}

	// Create a client for each device
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.AddEventHandler(ch.eventHandler)

	if config.Conf.Proxy.Enable {
		proxyData := ch.EnabledProxy(senderJidTypes.User)
		if proxyData != nil {
			client.SetProxy(http.ProxyURL(proxyData))
		}
	}

	// Connect the client synchronously
	err = client.Connect()
	if err != nil {
		log.Errorf("client.Connect(), got err : %v", err)
		err = errors.New("you are not logged in in this server before or just do the pair phone or scan qr")
		return err
	}
	// Add the client to the map
	StoreClientConcurrent(senderJidTypes.User, client)
	return nil
}

func (ch CommandHandler) HandleConnectBulkDevices(ctx context.Context, senderJidTypes []types.JID) error {
	// Retrieve all devices from the container
	devices, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("ch.Container.GetAllDevices, got err: %v", err)
		return err
	}

	// Filter devices based on sender JIDs
	var filteredDevices []*store.Device
	for _, singleDevice := range devices {
		for _, senderJid := range senderJidTypes {
			if strings.EqualFold(singleDevice.ID.User, senderJid.User) {
				filteredDevices = append(filteredDevices, singleDevice)
			}
		}
	}

	if len(filteredDevices) == 0 {
		log.Warn("No matching devices found for the given JIDs")
		return errors.New("no devices found matching the provided JIDs")
	}

	// Channel for error handling and WaitGroup for synchronization
	errCh := make(chan error, len(filteredDevices))
	var wg sync.WaitGroup

	for _, device := range filteredDevices {
		if device.ID.User == "" {
			log.Warn("Skipping device with empty user ID")
			continue
		}

		// Increment WaitGroup for the goroutine
		wg.Add(1)
		go func(device *store.Device) {
			defer wg.Done()

			// Create a client for the device
			clientLog := waLog.Stdout("Client", "ERROR", true)
			client := whatsmeow.NewClient(device, clientLog)
			client.AddEventHandler(ch.eventHandler)

			// Configure proxy if enabled
			if config.Conf.Proxy.Enable {
				proxyData := ch.EnabledProxy(device.ID.User)
				if proxyData != nil {
					client.SetProxy(http.ProxyURL(proxyData))
				}
			}

			// Connect the client
			err = client.Connect()
			if err != nil {
				log.Errorf("client.Connect() for device %s, got err: %v", device.ID.User, err)
				errCh <- fmt.Errorf("failed to connect device %s: %w", device.ID.User, err)
				return
			}

			// Store the client in a concurrent-safe way
			StoreClientConcurrent(device.ID.User, client)
		}(device)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errCh)

	// Check for any errors during connection
	if len(errCh) > 0 {
		var errList []string
		for e := range errCh {
			errList = append(errList, e.Error())
		}
		return fmt.Errorf("encountered errors: %s", strings.Join(errList, "; "))
	}

	log.Info("Successfully connected all devices")
	return nil
}

func (ch CommandHandler) HandleDisconnectSingleDevice(ctx context.Context, senderJidTypes types.JID) error {
	devices, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("ch.Container.GetDevice, got err : %v", err)
		err = errors.New("you are not logged in in this server before or just do the pair phone or scan qr")
		return err
	}

	var device *store.Device
	for _, singleDevice := range devices {
		if strings.EqualFold(singleDevice.ID.User, senderJidTypes.User) {
			device = singleDevice
		}
	}

	// Create a client for each device
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.AddEventHandler(ch.eventHandler)

	if config.Conf.Proxy.Enable {
		proxyData := ch.EnabledProxy(senderJidTypes.User)
		if proxyData != nil {
			client.SetProxy(http.ProxyURL(proxyData))
		}
	}

	// Connect the client synchronously
	client.Disconnect()
	log.Info("client.Disconnect()")

	// destroy the client of the map
	DeleteClientConcurrent(senderJidTypes.User)

	//destroy Proxy on runtime
	ch.ProxyManager.RemoveUser(senderJidTypes.User)

	return nil
}

func (ch CommandHandler) HandleDisconnectBulkDevices(ctx context.Context, senderJidTypes []types.JID) error {
	// Retrieve all devices from the container
	devices, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("ch.Container.GetAllDevices, got err: %v", err)
		return err
	}

	// Filter devices based on sender JIDs
	var filteredDevices []*store.Device
	for _, singleDevice := range devices {
		for _, senderJid := range senderJidTypes {
			if strings.EqualFold(singleDevice.ID.User, senderJid.User) {
				filteredDevices = append(filteredDevices, singleDevice)
			}
		}
	}

	if len(filteredDevices) == 0 {
		log.Warn("No matching devices found for the given JIDs")
		return errors.New("no devices found matching the provided JIDs")
	}

	// Channel for error handling and WaitGroup for synchronization
	errCh := make(chan error, len(filteredDevices))
	var wg sync.WaitGroup

	for _, device := range filteredDevices {
		if device.ID.User == "" {
			log.Warn("Skipping device with empty user ID")
			continue
		}

		// Increment WaitGroup for the goroutine
		wg.Add(1)
		go func(device *store.Device) {
			defer wg.Done()

			// Create a client for the device
			clientLog := waLog.Stdout("Client", "ERROR", true)
			client := whatsmeow.NewClient(device, clientLog)
			client.AddEventHandler(ch.eventHandler)

			// Configure proxy if enabled
			if config.Conf.Proxy.Enable {
				proxyData := ch.EnabledProxy(device.ID.User)
				if proxyData != nil {
					client.SetProxy(http.ProxyURL(proxyData))
				}
			}

			// Disconnect the client
			client.Disconnect()
			log.Infof("Disconnected device: %s", device.ID.User)

			// Remove the client from the concurrent-safe map
			DeleteClientConcurrent(device.ID.User)

			// Remove the proxy configuration for the device
			ch.ProxyManager.RemoveUser(device.ID.User)

		}(device)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errCh)

	// Check for any errors during disconnection
	if len(errCh) > 0 {
		var errList []string
		for e := range errCh {
			errList = append(errList, e.Error())
		}
		return fmt.Errorf("encountered errors during disconnection: %s", strings.Join(errList, "; "))
	}

	log.Info("Successfully disconnected all devices")
	return nil
}

func (ch CommandHandler) HandleLogoutSingleDevice(ctx context.Context, senderJidTypes types.JID) error {
	devices, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		log.Errorf("ch.Container.GetDevice, got err : %v", err)
		err = errors.New("you are not logged in in this server before or just do the pair phone or scan qr")
		return err
	}

	var device *store.Device
	for _, singleDevice := range devices {
		if strings.EqualFold(singleDevice.ID.User, senderJidTypes.User) {
			device = singleDevice
		}
	}

	// Create a client for each device
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.AddEventHandler(ch.eventHandler)

	if !client.IsConnected() {
		err = client.Connect()
		if err != nil {
			log.Errorf("client.Connect(), got err : %v", err)
			return err
		}
		StoreClientConcurrent(senderJidTypes.User, client)
		clientLoaded, ok := LoadClientConcurrent(senderJidTypes.User)
		if !ok || clientLoaded == nil {
			err = errors.New("cant connect to client")
			log.Errorf("client.Connect(), got err : %v", err)
			return err
		}
		err = clientLoaded.Logout(ctx)
		if err != nil {
			log.Errorf("client.Logout(), got err : %v", err)
			return err
		}
	} else {
		clientLoaded, ok := LoadClientConcurrent(senderJidTypes.User)
		if !ok || clientLoaded == nil {
			err = errors.New("cant connect to client")
			log.Errorf("client.Connect(), got err : %v", err)
			return err
		}
		err = clientLoaded.Logout(ctx)
		if err != nil {
			log.Errorf("client.Logout(), got err : %v", err)
			return err
		}
	}

	// Destroy client on the concurrent map
	DeleteClientConcurrent(senderJidTypes.User)

	//destroy Proxy on runtime
	ch.ProxyManager.RemoveUser(senderJidTypes.User)

	return nil
}

func (ch CommandHandler) HandleSpecificQRResponseCode(ctx context.Context, jid types.JID) (string, error) {
	devices, err := ch.Container.GetAllDevices(ctx)
	if err != nil {
		return "", err
	}

	var device *store.Device
	if len(devices) > 0 {
		for _, val := range devices {
			jidUser := strings.TrimSpace(jid.User)
			valIdUser := strings.TrimSpace(val.ID.User)

			if jidUser == valIdUser {
				device = val
			}
		}
	}

	if device == nil || device.ID.User == "" {
		device = ch.Container.NewDevice()
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)
	client.AddEventHandler(ch.eventHandler)

	if config.Conf.Proxy.Enable {
		proxyData := ch.EnabledProxy(jid.User)
		if proxyData != nil {
			client.SetProxy(http.ProxyURL(proxyData))
		}
	}

	// Connect the client synchronously
	if client.Store.ID == nil {
		qrChan, errGetQr := client.GetQRChannel(ctx)
		if errGetQr != nil {
			return "", errGetQr
		}
		err = client.Connect()
		if err != nil {
			return "", err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// Add the client to the map
				StoreClientConcurrent(jid.User, client)
				return evt.Code, nil
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			return "", err
		}
		// Add the client to the map
		StoreClientConcurrent(jid.User, client)
		return "", nil
	}
	return "", nil
}

func (ch CommandHandler) HandleSendImage(sender types.JID, JIDS []string, data []byte, captionMsg string) ([]primitive.Message, error) {
	return ch.handleSendMedia(sender, JIDS, data, captionMsg, whatsmeow.MediaImage, primitive.MediaImage, "")
}

func (ch CommandHandler) HandleSendDocument(sender types.JID, JIDS []string, fileName string, data []byte, captionMsg string) ([]primitive.Message, error) {
	return ch.handleSendMedia(sender, JIDS, data, captionMsg, whatsmeow.MediaDocument, primitive.MediaDocument, fileName)
}

func (ch CommandHandler) HandleSendVideo(sender types.JID, JIDS []string, data []byte, captionMsg string) ([]primitive.Message, error) {
	return ch.handleSendMedia(sender, JIDS, data, captionMsg, whatsmeow.MediaVideo, primitive.MediaVideo, "")
}

func (ch CommandHandler) HandleSendAudio(sender types.JID, JIDS []string, data []byte) ([]primitive.Message, error) {
	return ch.handleSendMedia(sender, JIDS, data, "", whatsmeow.MediaAudio, primitive.MediaAudio, "")
}

func (ch CommandHandler) HandleLoginAllDevices() {
	// Get all devices from the container
	devices, err := ch.Container.GetAllDevices(context.Background())
	if err != nil {
		log.Printf("Error retrieving devices: %v", err)
		return
	}

	// Create a buffered channel to collect errors from each goroutine
	errCh := make(chan error, len(devices))
	var wg sync.WaitGroup

	// Iterate over devices if any found
	for _, val := range devices {
		// Skip devices with no user ID
		if val.ID.User == "" {
			continue
		}

		// Increment the WaitGroup counter for each goroutine
		wg.Add(1)

		// Run the client connection in a separate goroutine
		go func(device *store.Device) {
			defer wg.Done() // Ensure the WaitGroup counter is decremented when this goroutine finishes

			// Initialize a new client with logging
			clientLog := waLog.Stdout("Client", "ERROR", true)
			client := whatsmeow.NewClient(device, clientLog)
			client.AddEventHandler(ch.eventHandler)

			// Check if the store is properly initialized
			if client.Store.ID == nil {
				errCh <- fmt.Errorf("skipping client: no store ID for user %s", device.ID.User)
				return
			}

			// Check if the client already connect skip
			if client.IsConnected() {
				errCh <- fmt.Errorf("skipping client: already connected %s", device.ID.User)
				return
			}

			if config.Conf.Proxy.Enable {
				proxyData := ch.EnabledProxy(device.ID.User)
				if proxyData != nil {
					client.SetProxy(http.ProxyURL(proxyData))
				}
			}

			// Connect the client
			if err = client.Connect(); err != nil {
				errCh <- fmt.Errorf("failed to connect client for user %s: %v", device.ID.User, err)
				return
			}

			// Store the client in the map concurrently
			StoreClientConcurrent(device.ID.User, client)
			log.Printf("Client connected for user %s", device.ID.User)
		}(val) // Pass the current value to the goroutine
	}

	// Start a separate goroutine to close the error channel once all the goroutines are done
	go func() {
		wg.Wait()    // Wait for all client connections to finish
		close(errCh) // Close the error channel after all connections have been processed
	}()

	// Process errors from the channel as they are received
	for err = range errCh {
		if err != nil {
			log.Printf("Error: %v", err) // Log each error received
		}
	}

	log.Println("HandleLoginAllDevices completed successfully")
}

func (ch CommandHandler) HandleDisconnectAllDevices() {
	// Get all devices from the container
	devices, err := ch.Container.GetAllDevices(context.Background())
	if err != nil {
		log.Printf("Error retrieving devices: %v", err)
		return
	}

	// Create a buffered channel to collect errors from each goroutine
	errCh := make(chan error, len(devices))
	var wg sync.WaitGroup

	// Iterate over devices if any found
	for _, val := range devices {
		// Skip devices with no user ID
		if val.ID.User == "" {
			continue
		}

		// Increment the WaitGroup counter for each goroutine
		wg.Add(1)

		// Run the client connection in a separate goroutine
		go func(device *store.Device) {
			defer wg.Done() // Ensure the WaitGroup counter is decremented when this goroutine finishes

			// Initialize a new client with logging
			clientLog := waLog.Stdout("Client", "ERROR", true)
			client := whatsmeow.NewClient(device, clientLog)
			client.AddEventHandler(ch.eventHandler)

			// Check if the client already connect skip
			if !client.IsConnected() {
				errCh <- fmt.Errorf("skipping client: already disconnected %s", device.ID.User)
				return
			}

			if config.Conf.Proxy.Enable {
				proxyData := ch.EnabledProxy(val.ID.User)
				if proxyData != nil {
					client.SetProxy(http.ProxyURL(proxyData))
				}
			}

			// Disconnect the client
			client.Disconnect()

			// Store the client in the map concurrently
			DeleteClientConcurrent(device.ID.User)
			log.Printf("Client disconnected for user %s", device.ID.User)
		}(val) // Pass the current value to the goroutine
	}

	// Start a separate goroutine to close the error channel once all the goroutines are done
	go func() {
		wg.Wait()    // Wait for all client connections to finish
		close(errCh) // Close the error channel after all connections have been processed
	}()

	// Process errors from the channel as they are received
	for err = range errCh {
		if err != nil {
			log.Printf("Error: %v", err) // Log each error received
		}
	}

	log.Println("HandleDisconnectAllDevices completed successfully")
}

// EnabledProxy returns the assigned proxy URL for a user, using proxy.Manager directly.
func (ch CommandHandler) EnabledProxy(senderJID string) *url.URL {
	// Try to get existing Proxy struct from Manager
	p, ok := ch.ProxyManager.GetUser(senderJID)
	if !ok {
		// No existing assignment, create one
		var err error
		p, err = ch.ProxyManager.AddUser(senderJID)
		if err != nil {
			log.Printf("ProxyManager.AddUser error: %v", err)
			return nil
		}
	}

	// Build URL
	proxyURL := &url.URL{
		Scheme: "socks5",
		User:   url.UserPassword(p.Username, p.Password),
		Host:   fmt.Sprintf("%s:%s", p.BaseURL, p.Port),
	}

	return proxyURL
}

func (ch CommandHandler) createImageMessage(uploaded whatsmeow.UploadResponse, data []byte, captionMsg string) waE2E.Message {
	return waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			URL:           proto.String(uploaded.URL),
			Mimetype:      proto.String(http.DetectContentType(data)),
			Caption:       &captionMsg,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			DirectPath:    proto.String(uploaded.DirectPath),
		},
	}
}

func (ch CommandHandler) createVideoMessage(uploaded whatsmeow.UploadResponse, data []byte, captionMsg string) waE2E.Message {
	return waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			URL:           proto.String(uploaded.URL),
			Mimetype:      proto.String(http.DetectContentType(data)),
			Caption:       &captionMsg,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			DirectPath:    proto.String(uploaded.DirectPath),
		},
	}
}

func (ch CommandHandler) createAudioMessage(uploaded whatsmeow.UploadResponse, data []byte) waE2E.Message {
	return waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{ // Change ImageMessage to AudioMessage
			URL:           proto.String(uploaded.URL),
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			DirectPath:    proto.String(uploaded.DirectPath),
		},
	}
}

func (ch CommandHandler) createDocumentMessage(fileName string, uploaded whatsmeow.UploadResponse, data []byte, captionMsg string) waE2E.Message {
	return waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			FileName:      proto.String(fileName),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
			Title:         proto.String(fmt.Sprintf("%s%s", "document", filepath.Ext(uploaded.URL))),
			Caption:       &captionMsg,
		},
	}
}

func (ch CommandHandler) handleSendMedia(sender types.JID, JIDS []string, data []byte, captionMsg string, mediaType whatsmeow.MediaType, mediaCategory string, fileName string) ([]primitive.Message, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sliceM []primitive.Message
	var errs []error

	// modern random API (Go 1.22+)
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()<<1)))

	for _, jid := range JIDS {
		wg.Add(1)
		go func(jid string) {
			defer wg.Done()

			// add random delay between 100ms–2000ms (2 seconds)
			delay := time.Duration(rng.IntN(1901)+100) * time.Millisecond
			time.Sleep(delay)

			recipient, ok := utils.ParseJID(jid)
			if !ok {
				mu.Lock()
				errs = append(errs, fmt.Errorf("invalid JID: %s", jid))
				mu.Unlock()
				return
			}

			client, ok := LoadClientConcurrent(sender.User)
			if !ok || client == nil {
				log.Printf("Client not found for user: %v", sender.User)
				return
			}

			uploaded, err := client.Upload(context.Background(), data, mediaType)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to upload file: %v", err))
				mu.Unlock()
				return
			}

			messageID := client.GenerateMessageID()

			var msg waE2E.Message
			switch mediaCategory {
			case primitive.MediaImage:
				msg = ch.createImageMessage(uploaded, data, captionMsg)
			case primitive.MediaDocument:
				msg = ch.createDocumentMessage(fileName, uploaded, data, captionMsg)
			case primitive.MediaVideo:
				msg = ch.createVideoMessage(uploaded, data, captionMsg)
			case primitive.MediaAudio:
				msg = ch.createAudioMessage(uploaded, data)
			default:
				mu.Lock()
				errs = append(errs, fmt.Errorf("unsupported media category: %s", mediaCategory))
				mu.Unlock()
				return
			}

			resp, err := client.SendMessage(context.Background(), recipient, &msg, whatsmeow.SendRequestExtra{ID: messageID})
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("error sending %s message: %v", mediaCategory, err))
				mu.Unlock()
				return
			}

			err = client.MarkRead([]types.MessageID{resp.ID}, time.Now(), recipient, sender)
			if err != nil {
				log.Printf("Error marking message as read: %v", err)
			}

			mu.Lock()
			sliceM = append(sliceM, primitive.Message{
				MessageID: resp.ID,
				Jid:       recipient.String(),
				Type:      mediaCategory,
				Body:      captionMsg,
				Sent:      true,
				FileName:  fileName,
			})
			mu.Unlock()
		}(jid)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	return sliceM, nil
}

func (ch CommandHandler) eventHandler(evt interface{}) {
	// TODO add action if something like the message has been delivered or has been readed.
}
