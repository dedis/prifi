package utils

import (
	prifilog "github.com/dedis/prifi/prifi-lib/log"
	"go.dedis.ch/onet/v3/log"
	"math"
	"strconv"
	"time"
)

// PCAPReceivedPacket represents a PCAP that was transmitted through Prifi and received at the relay
type PCAPReceivedPacket struct {
	ID              uint32
	clientID        uint16
	ReceivedAt      uint64
	SentAt          uint64
	Delay           uint64
	DataLen         uint32
	IsFinalFragment bool
}

// PCAPLog is a collection of PCAPReceivedPackets
type PCAPLog struct {
	reportID        int
	receivedPackets []*PCAPReceivedPacket
	nextReport      time.Time
	period          time.Duration
}

// Returns an instantiated PCAPLog
func NewPCAPLog() *PCAPLog {
	p := &PCAPLog{
		reportID:        0,
		receivedPackets: make([]*PCAPReceivedPacket, 0),
		period:          time.Duration(5) * time.Second,
		nextReport:      time.Now(),
	}
	return p
}

// should be called with the received pcap packet
func (pl *PCAPLog) ReceivedPcap(ID uint32, clientID uint16, frag bool, tsSent uint64, tsExperimentStart uint64, dataLen uint32) {

	if pl.receivedPackets == nil {
		pl.receivedPackets = make([]*PCAPReceivedPacket, 0)
	}

	receptionTime := uint64(prifilog.MsTimeStampNow()) - tsExperimentStart

	if receptionTime < 0 {
		receptionTime = 0
	}

	//log.Lvl1("Received PCAP", ID, "from client", clientID, "at time", receptionTime, "and it was sent at", tsSent, "so diff", receptionTime-tsSent)

	p := &PCAPReceivedPacket{
		ID:              ID,
		clientID:        clientID,
		ReceivedAt:      receptionTime,
		SentAt:          tsSent,
		Delay:           receptionTime - tsSent,
		DataLen:         dataLen,
		IsFinalFragment: frag,
	}

	pl.receivedPackets = append(pl.receivedPackets, p)

	now := time.Now()
	if now.After(pl.nextReport) {
		pl.Print()
		pl.nextReport = now.Add(pl.period)
	}
}

// prints current statistics for the pcap logger
func (pl *PCAPLog) Print() {

	totalPackets := 0
	totalUniquePackets := 0
	totalFragments := 0

	//compute min max and other stats
	delaysSum := uint64(0)
	delayMax := uint64(0)
	for _, v := range pl.receivedPackets {
		totalPackets++
		if v.IsFinalFragment {
			totalUniquePackets++
		} else {
			totalFragments++
		}

		delaysSum += v.Delay

		if v.Delay > delayMax {
			delayMax = v.Delay
		}
	}

	delayMean := float64(delaysSum) / float64(totalPackets)

	//now compute variance
	variance := float64(0)
	for _, v := range pl.receivedPackets {
		variance += (float64(v.Delay) - delayMean) * (float64(v.Delay) - delayMean)
	}

	variance = variance / float64(totalPackets)

	//compute stddev
	stddev := math.Sqrt(variance)

	log.Lvl1("PCAPLog (", pl.reportID, "): ", totalFragments, "fragments,", totalUniquePackets, "final,", totalPackets, "fragments+final; mean",
		math.Ceil(delayMean*100)/100, "ms, stddev", math.Ceil(stddev*100)/100, "max", math.Ceil(float64(delayMax)*100)/100, "ms")

	individualReports := make(map[uint16]string)
	for _, v := range pl.receivedPackets {
		if _, ok := individualReports[v.clientID]; !ok {
			individualReports[v.clientID] = ""
		}
		individualReports[v.clientID] += strconv.Itoa(int(v.Delay)) + ";"
	}

	for k, v := range individualReports {
		log.Lvl1("PCAPLog-individuals (", pl.reportID, "): client ", k, ":", v)
	}
	pl.reportID++
	pl.receivedPackets = make([]*PCAPReceivedPacket, 0)
}
