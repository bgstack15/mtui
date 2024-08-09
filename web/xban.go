package web

import (
	"encoding/json"
	"fmt"
	"mtui/bridge"
	"mtui/types"
	"mtui/types/command"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func SendLuaResponse(w http.ResponseWriter, err error, lr *command.LuaResponse) {
	if err != nil {
		SendError(w, 500, err)
	} else if !lr.Success {
		SendError(w, 500, fmt.Errorf(lr.Message))
	} else {
		Send(w, lr.Result, nil)
	}
}

func (a *Api) GetBanDBStatus(w http.ResponseWriter, r *http.Request) {
	req := &command.LuaRequest{
		Code: `
			local banned = 0
			local total = 0
			for _, entry in ipairs(xban.db) do
				total = total + 1
				if entry.banned then
					banned = banned + 1
				end
			end
			return { total = total, banned = banned }
		`,
	}
	resp := &command.LuaResponse{}
	err := a.app.Bridge.ExecuteCommand(command.COMMAND_LUA, req, resp, time.Second*5)
	SendLuaResponse(w, err, resp)
}

func (a *Api) GetBannedRecords(w http.ResponseWriter, r *http.Request) {
	req := &command.LuaRequest{
		Code: `
			local banned = {}
			for _, entry in ipairs(xban.db) do
				if entry.banned then
					table.insert(banned, entry)
				end
			end
			return banned
		`,
	}
	resp := &command.LuaResponse{}
	err := a.app.Bridge.ExecuteCommand(command.COMMAND_LUA, req, resp, time.Second*5)
	SendLuaResponse(w, err, resp)
}

func (a *Api) GetBanRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	e, err := a.app.GetOnlineXBanEntry(vars["playername"])
	Send(w, e, err)
}

func (a *Api) BanPlayer(w http.ResponseWriter, r *http.Request) {
	claims, err := a.GetClaims(r)
	if err != nil {
		SendError(w, 500, err)
		return
	}

	banr := &types.XBanRequest{}
	err = json.NewDecoder(r.Body).Decode(banr)
	if err != nil {
		SendError(w, 500, err)
		return
	}

	req := &command.LuaRequest{
		Code: fmt.Sprintf("return xban.ban_player('%s', '%s', nil, '%s')",
			bridge.SanitizeLuaString(banr.Playername),
			bridge.SanitizeLuaString(claims.Username),
			bridge.SanitizeLuaString(banr.Reason),
		),
	}
	resp := &command.LuaResponse{}
	err = a.app.Bridge.ExecuteCommand(command.COMMAND_LUA, req, resp, time.Second*5)
	SendLuaResponse(w, err, resp)
}

func (a *Api) TempBanPlayer(w http.ResponseWriter, r *http.Request) {
	claims, err := a.GetClaims(r)
	if err != nil {
		SendError(w, 500, err)
		return
	}

	banr := &types.XBanRequest{}
	err = json.NewDecoder(r.Body).Decode(banr)
	if err != nil {
		SendError(w, 500, err)
		return
	}

	req := &command.LuaRequest{
		Code: fmt.Sprintf("return xban.ban_player('%s', '%s', %d, '%s')",
			bridge.SanitizeLuaString(banr.Playername),
			bridge.SanitizeLuaString(claims.Username),
			banr.Time,
			bridge.SanitizeLuaString(banr.Reason),
		),
	}
	resp := &command.LuaResponse{}
	err = a.app.Bridge.ExecuteCommand(command.COMMAND_LUA, req, resp, time.Second*5)
	SendLuaResponse(w, err, resp)
}

func (a *Api) UnbanPlayer(w http.ResponseWriter, r *http.Request) {
	claims, err := a.GetClaims(r)
	if err != nil {
		SendError(w, 500, err)
		return
	}

	banr := &types.XBanRequest{}
	err = json.NewDecoder(r.Body).Decode(banr)
	if err != nil {
		SendError(w, 500, err)
		return
	}

	req := &command.LuaRequest{
		Code: fmt.Sprintf("return xban.unban_player('%s', '%s')",
			bridge.SanitizeLuaString(banr.Playername),
			bridge.SanitizeLuaString(claims.Username),
		),
	}
	resp := &command.LuaResponse{}
	err = a.app.Bridge.ExecuteCommand(command.COMMAND_LUA, req, resp, time.Second*5)
	SendLuaResponse(w, err, resp)
}

func (a *Api) CleanupBanDB(w http.ResponseWriter, r *http.Request) {
	req := &command.LuaRequest{
		Code: `
			local db = xban.db
			local old_count = #db
			local i = 1
			while i <= #db do
				if not db[i].banned then
					-- not banned, remove from db
					table.remove(db, i)
				else
					-- banned, hold entry back
					i = i + 1
				end
			end
			return {
				removed = (old_count - #db),
				retained = #db
			}
		`,
	}
	resp := &command.LuaResponse{}
	err := a.app.Bridge.ExecuteCommand(command.COMMAND_LUA, req, resp, time.Second*5)
	SendLuaResponse(w, err, resp)
}
