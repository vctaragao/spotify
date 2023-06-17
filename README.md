# Spotify Architeture - My Simplified Implementation

## Limitations

- Unlinke the original Spotify implementation my audio files will only one type of enconding. So It wont be necessary to filter out the peer how dont follow the same audio enconding that the searching peer use.
- Cached audio on the peers will only last for 30s insted of using the LRC policy
- RFC 5681 will not be modified
- Will not implement “emergency mode” It will just stop upload if a stutter have accoured
- No time limitation for oustanding peer requests
- No latency adjustment or Markovian implementation
- Peers wil not remenber the most recent searchs
- Client max number of peers connected wil only have a hard limit of 60
- For neighbor selection it will disconnect if after 30s no byte been transmited through the connection
- There will be no priority in the peer to peer conection, for now, it will just follow the First Come First Serve.
- Not NAT Traversal workaround

![image](https://github.com/vctaragao/spotify/assets/26884793/adeb54ef-bcb3-4ca3-96f6-dde028040411)

## Architeture Entities Responsabilities and Behaviours

- **********************Main Server**********************
    - Holds the original music files
    - Can respond with music in parts, very similar to a pagination funcionality
    - It will use Json Webtokens to autenticate the clients
- Tracker
    - Will hold the client to music mapping
    - Will have and enpoint to register that a certain client has a certain music track to distribute
    - Will have an endpoint that will return the 10 peers that have an specific desired track
- Client
    - Will have a Cache to store the played tracks
    - Will have a Buffer to hold the track data for the current playing music and prefetching
    - Will serve the overlay network with it’s fully cached tracks
    - Will consume the overlay network to download the track data that is current playing or prefetching

## Bussiness Rules

- The TCP Connections will be Keep-Alive conections. They will only terminate when desired, not when a data transfer is finished.
- Cache
    - The cache size will be 100MB in space
- Peer to Peer
    - A Client can only upload to a max of 4 different peers at the sime time
    - A Client will have a hard max connection limit to other peers of 60 at the same time
    - It can dowload from up to 5 peers simultaneously.
    - A track serverd through the peer-to-peer network will be split into 16KB chuncks
- Buffer
    - Will have no limit

## Bussiness Logic

- [ ]  When starting to play a track by _random access_ the Client will make a request to the server for the initial 15s of the track and, simultanesly, make a request to It’s Neighbors it the overlay network so that It can know wich peers to connect to to start downloading the track from the peers and not the server
- [ ]  When downloading a track from a peer It will always try to download the track completally
- [ ]  If the play-out buffer has less than 10s of music remain It will stop tring to download from the overlay peer-to-peer network and download the next 15s from the server.
- [ ]  If the play-out buffer hit a stutter than all upload should stop and the client only focus on the download of the current file from the server.
- [ ]  A Client connection to a peer will terminate if no data has been trasmited through the connection after 30s
- [ ]  The Client will start to prefetch a track in the overlay network if the current track has 30s or less until its finished. If the next track dosent have at least 15s prefetched when the current track has 10s or less until its finished, then the clint will prefetch the next track from the server insted of the peer-to-peer overlay network.
- [ ]  If a buffer underrun occurs in a track, the Spotify client pauses playback at that point
