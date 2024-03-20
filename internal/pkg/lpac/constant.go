package lpac

const (
	CommandStdioAPDU     = "apdu"
	CommandStdioLPA      = "lpa"
	CommandStdioProgress = "progress"

	CommandAPDUFuncConnect         = "connect"
	CommandAPDUFuncDisconnect      = "disconnect"
	CommandAPDUOpenLogicalChannel  = "logic_channel_open"
	CommandAPDUCloseLogicalChannel = "logic_channel_close"
	CommandAPDUFuncTransmit        = "transmit"
)

var ProgressMessages = map[string]string{
	"es10b_get_euicc_challenge_and_info":   "Getting euicc challenge and info",
	"es10b_retrieve_notifications_list":    "Retrieving notifications list",
	"es10b_authenticate_server":            "Authenticating server",
	"es10a_get_euicc_configured_addresses": "Getting euicc configured addresses",
	"es10b_get_euicc_info":                 "Getting euicc info",
	"es9p_initiate_authentication":         "Initiating authentication",
	"es9p_authenticate_client":             "Authenticating client",
	"es9p_handle_notification":             "Handling notification",
	"es9p_get_bound_profile_package":       "Getting bound profile package",
	"es10b_load_bound_profile_package":     "Loading bound profile package",
	"es10b_prepare_download":               "Preparing download",
}
