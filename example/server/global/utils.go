/**
 * Created: 2019/4/25 0025
 * @author: Jason
 */

package global


func GetAuthorizedData(srvType int, srvName string, key string) []byte {
	d := Authorized{
		Type: int32(srvType),
		Name: srvName,
		Key:  key,
	}
	return GetJSONMgr().MarshalData(&d)
}
