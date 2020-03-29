/**
 * Created: 2020/3/25
 * @author: Jason
 */

package msg

const (
	ST_NULL   = iota
	ST_GATE
	ST_GAME
	ST_CENTER
	s_MAX
)

const (
	RESULT_INVALID      = -1
	RESULT_OK           = iota
	RESULT_FAILED
	RESULT_DISCONN
	RESULT_SYNC_SESSION
	RESULT_MAX
)
