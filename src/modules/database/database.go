package database

import (
	"INServer/src/common/dbobj"
	"INServer/src/common/global"
	"INServer/src/common/logger"
	"INServer/src/common/uuid"
	"INServer/src/dao"
	"INServer/src/modules/node"
	"INServer/src/proto/data"
	"INServer/src/proto/db"
	"INServer/src/proto/msg"

	"github.com/gogo/protobuf/proto"
)

var Instance *Database

type (
	Database struct {
		DB                *dbobj.DBObject
		roleSummary       map[string]*data.RoleSummaryData
		roleSummaryByName map[string]*data.RoleSummaryData
		roles             map[string]*data.Role
		players           map[string]*data.Player
	}
)

func New() *Database {
	d := new(Database)
	d.roleSummary = make(map[string]*data.RoleSummaryData)
	d.roleSummaryByName = make(map[string]*data.RoleSummaryData)
	d.roles = make(map[string]*data.Role)
	d.players = make(map[string]*data.Player)
	d.DB = dbobj.New()
	d.DB.Open(global.ServerConfig.DatabaseConfig.Database, global.DatabaseSchema)
	d.loadAllRoleSummaryData()
	return d
}

func (d *Database) Start() {
	node.Instance.Net.Listen(msg.Command_LD_CREATE_PLAYER_REQ, d.onCreatePlayerReq)
	node.Instance.Net.Listen(msg.Command_GD_LOAD_PLAYER_REQ, d.onLoadPlayerReq)
	node.Instance.Net.Listen(msg.Command_GD_RELEASE_PLAYER_NTF, d.onReleasePlayerNtf)
	node.Instance.Net.Listen(msg.Command_GD_CREATE_ROLE_REQ, d.onCreateRoleReq)
	node.Instance.Net.Listen(msg.Command_GD_LOAD_ROLE_REQ, d.onLoadRoleReq)
}

func (d *Database) onCreatePlayerReq(header *msg.MessageHeader, buffer []byte) {
	resp := &msg.CreatePlayerResp{}
	defer node.Instance.Net.Responce(header, resp)
	message := &msg.CreatePlayerReq{}
	err := proto.Unmarshal(buffer, message)
	if err != nil {
		logger.Debug(err)
		return
	}
	player := &data.Player{}
	serializedData, err := proto.Marshal(player)
	if err != nil {
		logger.Debug(err)
		return
	}
	dbplayer := &db.DBPlayer{
		UUID:           message.PlayerUUID,
		SerializedData: serializedData,
	}
	err = dao.PlayerInsert(d.DB, dbplayer)
	if err != nil {
		logger.Debug(err)
		return
	}
	resp.Success = true
}

func (d *Database) onLoadPlayerReq(header *msg.MessageHeader, buffer []byte) {
	resp := &msg.LoadPlayerResp{}
	defer node.Instance.Net.Responce(header, resp)
	message := &msg.LoadPlayerReq{}
	err := proto.Unmarshal(buffer, message)
	if err != nil {
		logger.Debug(err)
		return
	}
	player, ok := d.players[message.PlayerUUID]
	if ok {
		resp.Success = true
		resp.Player = player
	} else {
		dbplayer, err := dao.PlayerQuery(d.DB, message.PlayerUUID)
		if err != nil {
			logger.Debug(err)
			return
		}
		player := &data.Player{}
		err = proto.Unmarshal(dbplayer.SerializedData, player)
		if err != nil {
			logger.Debug(err)
			return
		}
		resp.Success = true
		resp.Player = player
		d.players[message.PlayerUUID] = player
	}
}
func (d *Database) onReleasePlayerNtf(header *msg.MessageHeader, buffer []byte) {
	message := &msg.ReleasePlayerNtf{}
	err := proto.Unmarshal(buffer, message)
	if err != nil {
		logger.Debug(err)
		return
	}
	if _, ok := d.players[message.PlayerUUID]; ok {
		delete(d.players, message.PlayerUUID)
	}
}

func (d *Database) onCreateRoleReq(header *msg.MessageHeader, buffer []byte) {
	resp := &msg.CreateRoleResp{}
	defer node.Instance.Net.Responce(header, resp)
	message := &msg.CreateRoleReq{}
	err := proto.Unmarshal(buffer, message)
	if err != nil {
		logger.Debug(err)
		return
	}
	if player, ok := d.players[message.PlayerUUID]; ok {
		if _, ok := d.roleSummaryByName[message.RoleName]; ok {
			return
		}
		roleUUID := uuid.New()
		roleSummaryData := &data.RoleSummaryData{
			Name: message.RoleName,
			Zone: message.Zone,
		}
		summaryData, err := proto.Marshal(roleSummaryData)
		if err != nil {
			return
		}
		roleOnlineData := &data.RoleOnlineData{}
		onlineData, err := proto.Marshal(roleOnlineData)
		if err != nil {
			return
		}
		dbrole := &db.DBRole{
			UUID:        roleUUID,
			SummaryData: summaryData,
			OnlineData:  onlineData,
		}
		err = dao.RoleInsert(d.DB, dbrole)
		if err != nil {
			logger.Debug(err)
			return
		}
		player.RoleList = append(player.RoleList, roleSummaryData)
		d.roleSummaryByName[message.RoleName] = roleSummaryData
		d.roleSummary[roleUUID] = roleSummaryData
		resp.Success = true
		resp.RoleUUID = roleUUID
	} else {
		logger.Debug("Must Create Player Before Create Role!")
	}
}

func (d *Database) onLoadRoleReq(header *msg.MessageHeader, buffer []byte) {
	resp := &msg.LoadRoleResp{}
	defer node.Instance.Net.Responce(header, resp)
	message := &msg.LoadRoleReq{}
	err := proto.Unmarshal(buffer, message)
	if err != nil {
		logger.Debug(err)
		return
	}

	if roleSummary, ok := d.roleSummary[message.RoleUUID]; ok {
		onlineData, err := dao.RoleOnlineDataQuery(d.DB, message.RoleUUID)
		if err != nil {
			logger.Debug(err)
			return
		}
		roleOnline := &data.RoleOnlineData{}
		err = proto.Unmarshal(onlineData, roleOnline)
		if err != nil {
			logger.Debug(err)
			return
		}

		role := &data.Role{
			SummaryData: roleSummary,
			OnlineData:  roleOnline,
		}
		d.roles[message.RoleUUID] = role
	}
}

func (d *Database) loadAllRoleSummaryData() {
	roles := dao.AllRoleSummaryDataQuery(d.DB)
	for _, role := range roles {
		summary := &data.RoleSummaryData{}
		proto.Unmarshal(role.SummaryData, summary)
		d.roleSummary[role.UUID] = summary
	}

	for _, role := range d.roleSummary {
		d.roleSummaryByName[role.Name] = role
	}

	logger.Debug("加载所有角色的摘要数据成功")
}