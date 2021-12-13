package snowflake

import (
	"errors"
	"net"
	"sync"
	"time"
)

/**
雪花算法
	 41bit timestamp | 10 bit machineID ｜ 12 bit sequenceBits
*/

const (
	timestampBits = uint64(41)
	machineIDBits = uint64(10)
	sequenceBits = uint64(12)
	maxSequence  = 1<<sequenceBits - 1
	timeLeft     = uint8(22)
)

type Snowflake struct {
	mutex     *sync.Mutex
	StartTime int64
	LastStamp int64
	MachineID uint16
	Sequence  int64
}

func GenSnowflake() (uint64,error){
	f:=NewSnowflake()
	return f.NextID()
}


func NewSnowflake() *Snowflake {
	st := time.Date(2019, 4, 21, 0, 0, 0, 0, time.UTC).UnixNano() / 1e6
	mID, err := lower10BitPrivateIP()
	if err != nil {
		panic(err)
	}
	return &Snowflake{
		StartTime: st,
		MachineID: mID,
		LastStamp: 0,
		Sequence:  0,
		mutex: new(sync.Mutex),
	}
}

func (sf *Snowflake) getMilliSeconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func (sf *Snowflake) NextID() (uint64, error) {
	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	return sf.nextID()
}

func (sf *Snowflake) nextID() (uint64, error) {
	timeStamp := sf.getMilliSeconds()
	if timeStamp < sf.LastStamp {
		return 0, errors.New("time is moving backwards")
	}
	if sf.LastStamp == timeStamp {
		sf.Sequence = (sf.Sequence + 1) & maxSequence
		if sf.Sequence == 0 {
			for timeStamp <= sf.LastStamp {
				timeStamp = sf.getMilliSeconds()
			}
		}
	} else {
		sf.Sequence = 0
	}
	sf.LastStamp = timeStamp

	id := ((timeStamp - sf.StartTime) << timeLeft) |
		int64(sf.MachineID<<sequenceBits) |
		sf.Sequence

	return uint64(id), nil
}

func lower10BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return (uint16(ip[2])<<14)>>6 + uint16(ip[3]), nil
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}
