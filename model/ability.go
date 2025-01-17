package model

import (
	"fmt"
	"one-api/common"
	"strings"
)

type Ability struct {
	Group             string `json:"group" gorm:"type:varchar(32);primaryKey;autoIncrement:false"`
	Model             string `json:"model" gorm:"primaryKey;autoIncrement:false"`
	ChannelId         int    `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled           bool   `json:"enabled"`
	AllowStreaming    int    `json:"allow_streaming" gorm:"default:1"`
	AllowNonStreaming int    `json:"allow_non_streaming" gorm:"default:1"`
}

func GetRandomSatisfiedChannel(group string, model string, stream bool) (*Channel, error) {
	ability := Ability{}
	var err error = nil

	cmd := "`group` = ? and model = ? and enabled = 1"

	if common.UsingPostgreSQL {
		// Make cmd compatible with PostgreSQL
		cmd = "\"group\" = ? and model = ? and enabled = true"
	}

	if stream {
		cmd += fmt.Sprintf(" and allow_streaming = %d", common.ChannelAllowStreamEnabled)
	} else {
		cmd += fmt.Sprintf(" and allow_non_streaming = %d", common.ChannelAllowNonStreamEnabled)
	}

	if common.UsingSQLite || common.UsingPostgreSQL {
		err = DB.Where(cmd, group, model).Order("RANDOM()").Limit(1).First(&ability).Error
	} else {
		err = DB.Where(cmd, group, model).Order("RAND()").Limit(1).First(&ability).Error
	}
	if err != nil {
		return nil, err
	}
	channel := Channel{}
	channel.Id = ability.ChannelId
	err = DB.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}

func (channel *Channel) AddAbilities() error {
	models_ := strings.Split(channel.Models, ",")
	groups_ := strings.Split(channel.Group, ",")
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			ability := Ability{
				Group:             group,
				Model:             model,
				ChannelId:         channel.Id,
				Enabled:           channel.Status == common.ChannelStatusEnabled,
				AllowStreaming:    channel.AllowStreaming,
				AllowNonStreaming: channel.AllowNonStreaming,
			}
			abilities = append(abilities, ability)
		}
	}
	return DB.Create(&abilities).Error
}

func (channel *Channel) DeleteAbilities() error {
	return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities() error {
	// A quick and dirty way to update abilities
	// First delete all abilities of this channel
	err := channel.DeleteAbilities()
	if err != nil {
		return err
	}
	// Then add new abilities
	err = channel.AddAbilities()
	if err != nil {
		return err
	}
	return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}
