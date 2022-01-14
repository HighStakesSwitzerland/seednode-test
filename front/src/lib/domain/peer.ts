export interface Peer {
  NodeInfo: NodeInfo
  LastValue: number
}

export interface NodeInfo {
  id: string,
  listen_addr: string,
  network: string,
  version: string,
  moniker: string,
}

export interface Seed {

}
