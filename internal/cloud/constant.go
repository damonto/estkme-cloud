package cloud

type Tag byte

const (
	TagMessageBox          Tag = 0x00
	TagManagement          Tag = 0x01
	TagDownloadProfile     Tag = 0x02
	TagProcessNotification Tag = 0x03
	TagReboot              Tag = 0xFB
	TagClose               Tag = 0xFC
	TagAPDULock            Tag = 0xFD
	TagAPDU                Tag = 0xFE
	TagAPDUUnlock          Tag = 0xFF
)

var KnownTags = []Tag{
	TagAPDU,
	TagAPDULock,
	TagAPDUUnlock,
	TagClose,
	TagDownloadProfile,
	TagProcessNotification,
	TagManagement,
	TagMessageBox,
	TagReboot,
}
