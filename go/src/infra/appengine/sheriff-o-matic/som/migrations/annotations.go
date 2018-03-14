package migrations

import (
  "net/http"

  "infra/appengine/sheriff-o-matic/som/handler"
  "infra/appengine/sheriff-o-matic/som/model"

  "go.chromium.org/gae/service/datastore"
  "go.chromium.org/luci/server/router"
)

// Attaches trees to annotations after the annotation schema was changed to
// include trees.
func AnnotationTreeWorker(ctx *router.Context) {
	c, w := ctx.Context, ctx.Writer

  q := datastore.NewQuery("Annotation")
  annotations := []*model.Annotation{}
	err := datastore.GetAll(c, q, &annotations)
  if err != nil {
    handler.ErrStatus(c, w, http.StatusInternalServerError, err.Error())
    return
  }

  q = datastore.NewQuery("AlertJSON")
  alerts := []*model.AlertJSON{}
  err = datastore.GetAll(c, q, &alerts)
  if err != nil {
    handler.ErrStatus(c, w, http.StatusInternalServerError, err.Error())
    return
  }

  alertMap := make(map[string]*model.AlertJSON)
  for _, alert := range alerts {
    alertMap[alert.ID] = alert
  }

  for _, ann := range annotations {
    alert := alertMap[ann.Key]
    ann.Tree = alert.Tree
  }

  err = datastore.Put(c, alerts)
  if err != nil {
    handler.ErrStatus(c, w, http.StatusInternalServerError, err.Error())
    return
  }

  w.Write([]byte("ok"))
}
