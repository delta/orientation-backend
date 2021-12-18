package models

type Room struct {
	ID   int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name string `gorm:"column:name;not null;unique"`
}

func (Room) TableName() string {
	return "Room"
}

func GetAllRooms() ([]string, error) {
	var rooms []string

	if err := db.Model(&Room{}).Select("name").Find(&rooms).Error; err != nil {
		return nil, err
	}

	return rooms, nil
}
