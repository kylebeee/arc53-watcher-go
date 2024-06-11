package server

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kylebeee/arc53-watcher-go/db"
	"github.com/kylebeee/arc53-watcher-go/db/compound"
	"github.com/kylebeee/arc53-watcher-go/errors"
	"github.com/kylebeee/arc53-watcher-go/providers"
)

func (s *Arc53WatcherServer) handleHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"ok": true,
		})
	}
}

func (s *Arc53WatcherServer) handleGetARC53Data() gin.HandlerFunc {
	const op errors.Op = "handleGetARC53Data"

	type request struct {
		AppID string `uri:"appID" binding:"required"`
	}

	type response struct {
		*compound.Community `json:"community,omitempty"`
		Error               string `json:"error,omitempty"`
	}

	return func(c *gin.Context) {
		var (
			req  request
			resp response
			err  error
		)

		err = c.ShouldBindUri(&req)
		if err != nil {
			c.JSON(400, gin.H{
				"ok":    false,
				"error": err.Error(),
			})
			return
		}

		appID, err := strconv.ParseUint(req.AppID, 10, 64)
		if err != nil {
			err = errors.E(op, err)
			fmt.Print(err)
			resp.Error = "bad request"
			c.JSON(400, resp)
			return
		}

		resp.Community, err = compound.GetCommunity(s.DB, appID)
		if err != nil && !db.ErrNoRows(err) {
			err = errors.E(op, err)
			fmt.Print(err)
			resp.Error = "internal server error"
			c.JSON(500, resp)
			return
		} else if db.ErrNoRows(err) {
			resp.Error = "not found"
			c.JSON(404, resp)
			return
		}
	}
}

func (s *Arc53WatcherServer) handleSyncByProviderID() gin.HandlerFunc {
	const op errors.Op = "handleSyncByProviderID"

	type request struct {
		ProviderType string `uri:"providerType" binding:"required"`
		AppID        string `uri:"appID" binding:"required"`
	}

	type response struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	return func(c *gin.Context) {
		var (
			req  request
			resp response
			err  error
		)

		err = c.ShouldBindUri(&req)
		if err != nil {
			c.JSON(400, gin.H{
				"ok":    false,
				"error": err.Error(),
			})
			return
		}

		providerTypeMap := map[string]providers.ProviderType{}

		for i := range s.ProviderTypes {
			providerTypeMap[s.ProviderTypes[i].Type()] = s.ProviderTypes[i]
		}

		provider, ok := providerTypeMap[req.ProviderType]
		if !ok {
			resp.Error = "provider not found"
			c.JSON(404, resp)
			return
		}

		appID, err := strconv.ParseUint(req.AppID, 10, 64)
		if err != nil {
			err = errors.E(op, err)
			fmt.Print(err)
			resp.Error = "bad request"
			c.JSON(400, resp)
			return
		}

		c.JSON(200, gin.H{
			"ok": true,
		})

		go provider.Process(appID)
	}
}
