package pluginproxy

import (
//	"bytes"
	"fmt"
	"io/ioutil"
//	"net/http/httputil"
	"net/http"
//    "reflect"
//	"github.com/grafana/grafana/pkg/log"
	"strings"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/bus"
	//"github.com/grafana/grafana/pkg/services/sqlstore"
)

type myTransport struct {
	proxy *DataSourceProxy
}
// Crée un filtre elasticsearch pour l'utilisateur connecté
func (t *myTransport) getFilter() (string) {

	user:=t.proxy.ctx.SignedInUser
	teamQuery := m.GetTeamsByUserQuery{OrgId: user.OrgId, UserId: user.UserId}
	bus.Dispatch(&teamQuery)

	var teams = teamQuery.Result
	var query = "{ \"bool\":{ \"should\": ["
	for i := 0; i < len(teams); i++ {
		fmt.Println(string(teams[i].Name))
		query = query + "{\"term\": {\"project.name.keyword\": \"" + teams[i].Name +"\"}}"
        if (i<len(teams) -1) {
        	query = query + ","
        }
    }
    query = query + "]}}"
	fmt.Println(query)
	return query

}
func (t *myTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// Lecture du body de la requête original
	buf, _ := ioutil.ReadAll(request.Body)
	var body string= string(buf)

	var filter = t.getFilter()
	// Insertion du filtre utilisateur dans le nouveau body
	var new_body = strings.Replace(body,"\"filter\":[", "\"filter\":[" + filter + ",",-1)
	//fmt.Println(new_body)

	// Remplacement du body de la requête originale
	request.Body = ioutil.NopCloser(strings.NewReader(new_body))

	// Mise à jour du content_length
	request.ContentLength = int64(len(new_body))
  
  	// Exécution de la méthode roundtrip par défaut
	response, err:= http.DefaultTransport.RoundTrip(request)
	if err != nil {
		print("\n\ncame in error resp here", err)
		return nil, err //Server is not reachable. Server not working
	}
	
	/*
	// Logging de la réponse
	response_body, err := httputil.DumpResponse(response, true)
	if err != nil {
		print("\n\nerror in dumb response")
		// copying the response body did not work
		return nil, err
	}
	log.Info("Response Body : ", string(body))
	*/
	return response, err
}
