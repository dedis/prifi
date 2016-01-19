package relay

import (
	"github.com/lbarman/prifi/config"
	"time"
	"net"
	"errors"
	"strconv"
	prifinet "github.com/lbarman/prifi/net"
	prifilog "github.com/lbarman/prifi/log"
)

func initiateRelayState(relayPort string, nTrustees int, nClients int, upstreamCellSize int, downstreamCellSize int, useDummyDataDown bool, reportingLimit int, trusteesHosts []string, useUDP bool) *RelayState {
	params := new(RelayState)
	
	params.Name               = "Relay"
	params.RelayPort          = relayPort
	params.UpstreamCellSize   = upstreamCellSize
	params.DownstreamCellSize = downstreamCellSize
	params.ReportingLimit     = reportingLimit
	params.UseUDP             = useUDP
	params.UseDummyDataDown   = useDummyDataDown

	//prepare the crypto parameters
	rand 	:= config.CryptoSuite.Cipher([]byte(params.Name))
	base	:= config.CryptoSuite.Point().Base()

	//generate own parameters
	params.privateKey       = config.CryptoSuite.Secret().Pick(rand)
	params.PublicKey        = config.CryptoSuite.Point().Mul(base, params.privateKey)

	params.nClients      = nClients
	params.nTrustees     = nTrustees
	params.trusteesHosts = trusteesHosts

	//sets the cell coder, and the history
	params.CellCoder = config.Factory()

	return params
}

func (relayState *RelayState) deepClone() *RelayState {
	newRelayState := new(RelayState)

	newRelayState.Name               = relayState.Name
	newRelayState.RelayPort          = relayState.RelayPort
	newRelayState.PublicKey          = relayState.PublicKey
	newRelayState.privateKey         = relayState.privateKey
	newRelayState.nClients           = relayState.nClients
	newRelayState.nTrustees          = relayState.nTrustees
	newRelayState.trusteesHosts      = make([]string, len(relayState.trusteesHosts))
	newRelayState.clients            = make([]prifinet.NodeRepresentation, len(relayState.clients))
	newRelayState.trustees           = make([]prifinet.NodeRepresentation, len(relayState.trustees))
	newRelayState.CellCoder          = config.Factory()
	newRelayState.MessageHistory     = relayState.MessageHistory
	newRelayState.UpstreamCellSize   = relayState.UpstreamCellSize
	newRelayState.DownstreamCellSize = relayState.DownstreamCellSize
	newRelayState.ReportingLimit     = relayState.ReportingLimit
	newRelayState.UseUDP 			 = relayState.UseUDP
	newRelayState.UseDummyDataDown   = relayState.UseDummyDataDown
	newRelayState.UDPBroadcastConn 	 = relayState.UDPBroadcastConn

	copy(newRelayState.trusteesHosts, relayState.trusteesHosts)

	for i:=0; i<len(relayState.clients); i++{
		newRelayState.clients[i].Id        = relayState.clients[i].Id
		newRelayState.clients[i].Conn      = relayState.clients[i].Conn
		newRelayState.clients[i].Connected = relayState.clients[i].Connected
		newRelayState.clients[i].PublicKey = relayState.clients[i].PublicKey
	}
	for i:=0; i<len(relayState.trustees); i++{
		newRelayState.trustees[i].Id        = relayState.trustees[i].Id
		newRelayState.trustees[i].Conn      = relayState.trustees[i].Conn
		newRelayState.trustees[i].Connected = relayState.trustees[i].Connected
		newRelayState.trustees[i].PublicKey = relayState.trustees[i].PublicKey
	}

	return newRelayState
}

func (relayState *RelayState) addNewClient(newClient prifinet.NodeRepresentation){
	relayState.nClients = relayState.nClients + 1
	relayState.clients  = append(relayState.clients, newClient)
}

func connectToTrusteeAsync(trusteeChan chan prifinet.NodeRepresentation, id int, host string, relayState *RelayState) {
		
	var err error = errors.New("empty")
	trustee := prifinet.NodeRepresentation{}

	for err != nil {
		trustee, err = connectToTrustee(id, host,relayState)

		if err != nil { 
			prifilog.Println(prifilog.RECOVERABLE_ERROR, "Failed to connect to trustee " + strconv.Itoa(id) + " host " + host + ", retrying...")
		}
	}
	
	trusteeChan <- trustee
}

func (relayState *RelayState) connectToAllTrustees() {

	defer prifilog.TimeTrack("relay", "connectToAllTrustees", time.Now())

	trusteeChan := make(chan prifinet.NodeRepresentation, relayState.nTrustees)

	//connect to all the trustees
	for i:= 0; i < relayState.nTrustees; i++ {
		go connectToTrusteeAsync(trusteeChan, i, relayState.trusteesHosts[i], relayState)
	}

	//wait for all the trustees to be connected
	i := 0
	for i < relayState.nTrustees {
		select {
			case trustee := <- trusteeChan:
				relayState.trustees = append(relayState.trustees, trustee)
				i++

			default:
				time.Sleep(10 * time.Millisecond)
		}
	}

	prifilog.Println(prifilog.INFORMATION, "Trustees connecting done, ", len(relayState.trustees), "trustees connected")
}

func (relayState *RelayState) disconnectFromAllTrustees() {
	defer prifilog.TimeTrack("relay", "disconnectToAllTrustees", time.Now())

	//disconnect to the trustees
	for i:= 0; i < len(relayState.trustees); i++ {
		relayState.trustees[i].Conn.Close()
	}
	relayState.trustees = make([]prifinet.NodeRepresentation, 0)
	prifilog.Println(prifilog.INFORMATION, "Trustees disonnecting done, ", len(relayState.trustees), "trustees disconnected")
}


func welcomeNewClients(newConnectionsChan chan net.Conn, newClientChan chan prifinet.NodeRepresentation, clientsUseUDP bool) {	
	newClientsToParse := make(chan prifinet.NodeRepresentation)

	for {
		select{
			//accept the TCP connection, and parse the parameters
			case newConnection := <-newConnectionsChan: 
				prifilog.Println(prifilog.INFORMATION, "welcomeNewClients : New connection is ready")
				go relayParseClientParams(newConnection, newClientsToParse, clientsUseUDP)
			
			//once client is ready (we have params+pk), forward to the other channel
			case newClient := <-newClientsToParse: 
				prifilog.Println(prifilog.INFORMATION, "welcomeNewClients : New client is ready")
				newClientChan <- newClient

			default: 
				time.Sleep(NEWCLIENT_CHECK_SLEEP_TIME) //todo : check this duration
				//prifilog.StatisticReport("relay", "NEWCLIENT_CHECK_SLEEP_TIME", "NEWCLIENT_CHECK_SLEEP_TIME")
		}
	}
}

func (relayState *RelayState) waitForDefaultNumberOfClients(newClientConnectionsChan chan prifinet.NodeRepresentation) {
	defer prifilog.TimeTrack("relay", "waitForDefaultNumberOfClients", time.Now())

	currentClients := 0

	prifilog.Printf(prifilog.INFORMATION, "Waiting for %d clients (on port %s)", relayState.nClients - currentClients, relayState.RelayPort)

	for currentClients < relayState.nClients {
		select{
				case newClient := <-newClientConnectionsChan: 
					relayState.clients = append(relayState.clients, newClient)
					currentClients += 1
					prifilog.Printf(prifilog.INFORMATION, "Waiting for %d clients (on port %s)", relayState.nClients - currentClients, relayState.RelayPort)
				default: 
					time.Sleep(100 * time.Millisecond)
					//prifilog.StatisticReport("relay", "SLEEP_100ms", "100ms")
		}
	}
	prifilog.Println(prifilog.INFORMATION, "Client connecting done, ", len(relayState.clients), "clients connected")
}

func (relayState *RelayState) excludeDisconnectedClients(){
	defer prifilog.TimeTrack("relay", "excludeDisconnectedClients", time.Now())

	//count the clients that disconnected
	nClientsDisconnected := 0
	for i := 0; i<len(relayState.clients); i++ {
		if !relayState.clients[i].Connected {
			prifilog.Println(prifilog.INFORMATION, "Relay Handler : Client ", i, " discarded, seems he disconnected...")
			nClientsDisconnected++
		}
	}

	//count the actual number of clients, and init the new state with the old parameters
	newNClients   := relayState.nClients - nClientsDisconnected

	//copy the connected clients
	newClients := make([]prifinet.NodeRepresentation, newNClients)
	j := 0
	for i := 0; i<len(relayState.clients); i++ {
		if relayState.clients[i].Connected {
			newClients[j] = relayState.clients[i]
			prifilog.Println(prifilog.INFORMATION, "Adding Client ", i, "who's not disconnected")
			j++
		}
	}

	relayState.clients = newClients
}