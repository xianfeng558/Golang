package data

import (
	"errors"

	"customer/internal/biz"
)

// CreateTrip 插入一条待接单的行程。价格单位是分。
func (cd *CustomerData) CreateTrip(customerID uint, origin, destination string, price int64) (*biz.TripOrder, error) {
	row := &biz.TripOrder{
		CustomerID:  customerID,
		Origin:      origin,
		Destination: destination,
		Price:       price,
		Status:      biz.TripWaitDriver,
	}
	if err := cd.data.Mdb.Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

// ListTrips 按乘客主键返回行程，新的在前。
func (cd *CustomerData) ListTrips(customerID uint) ([]biz.TripOrder, error) {
	rows := make([]biz.TripOrder, 0)
	err := cd.data.Mdb.Where("customer_id = ?", customerID).Order("id desc").Find(&rows).Error
	return rows, err
}

// TransitTrip 把行程改到 to。只有这一行的乘客可以改，并且必须符合状态机。
func (cd *CustomerData) TransitTrip(customerID, tripID uint, to string) (*biz.TripOrder, error) {
	row := &biz.TripOrder{}
	if err := cd.data.Mdb.First(row, tripID).Error; err != nil {
		return nil, err
	}
	if row.CustomerID != customerID {
		return nil, errors.New("行程不属于当前乘客")
	}
	if !biz.CanTransitTrip(row.Status, to) {
		return nil, errors.New("行程状态不允许这样变化")
	}
	if err := cd.data.Mdb.Model(row).Update("status", to).Error; err != nil {
		return nil, err
	}
	row.Status = to
	return row, nil
}
