package lpac

type Command string

const (
	CommandStdioAPDU               Command = "apdu"
	CommandStdioLPA                Command = "lpa"
	CommandStdioProgress           Command = "progress"
	CommandAPDUFuncConnect         Command = "connect"
	CommandAPDUFuncDisconnect      Command = "disconnect"
	CommandAPDUOpenLogicalChannel  Command = "logic_channel_open"
	CommandAPDUCloseLogicalChannel Command = "logic_channel_close"
	CommandAPDUFuncTransmit        Command = "transmit"
)

var HumanReadableSteps = map[string]string{
	"es9p_initiate_authentication":         "Initiating authentication",
	"es9p_authenticate_client":             "Authenticating client",
	"es9p_handle_notification":             "Handling notification",
	"es9p_get_bound_profile_package":       "Getting bound profile package",
	"es10a_get_euicc_configured_addresses": "Getting eUICC configured addresses",
	"es10b_get_euicc_challenge_and_info":   "Getting eUICC challenge and info",
	"es10b_retrieve_notifications_list":    "Retrieving notifications list",
	"es10b_authenticate_server":            "Authenticating server",
	"es10b_get_euicc_info":                 "Getting eUICC info",
	"es10b_load_bound_profile_package":     "Loading bound profile package, It may take a few minutes",
	"es10b_prepare_download":               "Preparing download",
}
