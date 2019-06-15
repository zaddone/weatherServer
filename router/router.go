package router
import(
	"io"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"github.com/gin-gonic/gin"
	"strings"
	//"bytes"
	"github.com/boltdb/bolt"
	"fmt"
	"github.com/zaddone/studySystem/request"
	//"github.com/unrolled/secure"
)
const(
	addrDB string = "addr.db"
	cityM string = "http://toy1.weather.com.cn/search?cityname="
	weatherUrl string = "http://d1.weather.com.cn/weather_index/"
)
var (
	db *bolt.DB
)
func GetCityID(city string,hand func(string)error) (err error) {
	ci := strings.Split(city,",")
	if len(ci) !=2 {
		return fmt.Errorf("%s",ci)
	}
	err = db.View(func(t *bolt.Tx)error{
		b := t.Bucket([]byte(addrDB))
		if b == nil {
			return fmt.Errorf("b=nil")
		}
		id := b.Get([]byte(city))
		if id != nil {
			return hand(string(id))
		}
		return fmt.Errorf("id=nil")
	})
	if err == nil {
		return nil
	}
	return request.ClientHttp_(cityM+url.QueryEscape(ci[1]),"GET",nil,nil,func(body io.Reader,st int)error{
		c,err := ioutil.ReadAll(body)
		if err != nil {
			return err
		}

		if st != 200 {
			return fmt.Errorf("%s",c)
		}
		var c_ interface{}
		err = json.Unmarshal(c[1:len(c)-1],&c_)
		//fmt.Println(string(c[1:len(c)-1]),c_,err)
		if err != nil {
			return err
		}
		for _,c__ := range c_.([]interface{}) {
			_cs := strings.Split((c__.(map[string]interface{}))["ref"].(string),"~")
			//fmt.Println(_cs)
			if strings.Contains(_cs[2],ci[1]) && strings.Contains(_cs[9],ci[0]){
				err = db.Batch(func(t *bolt.Tx)error{
					b,err := t.CreateBucketIfNotExists([]byte(addrDB))
					if err != nil {
						return err
					}
					return b.Put([]byte(city),[]byte(_cs[0]))
				})
				if err != nil {
					return err
				}
				return hand(_cs[0])
			}


		}
		return fmt.Errorf("%s",c)

	})
}
func init(){

	gin.SetMode(gin.ReleaseMode)
	var err error
	db,err = bolt.Open(addrDB,0600,nil)
	if err != nil {
		panic(err)
	}
	Router := gin.Default()
	Router.GET("/getweather",func(c *gin.Context){
		var msg string
		err = GetCityID(c.Query("city"),func(id string)error{
			return request.ClientHttp_(fmt.Sprintf("%s%s.html",weatherUrl,id),"GET",nil,http.Header{"Referer":[]string{"http://www.weather.com.cn/"}},func(body io.Reader,st int)error{
				c,err := ioutil.ReadAll(body)
				if err != nil {
					return err
				}
				if st != 200 {
					return fmt.Errorf("%s",c)
				}
				msg = string(c)
				return nil
			})
		})
		if err != nil {
			c.JSON(http.StatusFound,gin.H{"status":-1,"msg":err})
			return
		}
		c.JSON(http.StatusOK,gin.H{"status":1,"msg":msg})
		return
	})
	go Router.RunTLS(":443", "2301242_zaddone.com.pem", "2301242_zaddone.com.key")

}
