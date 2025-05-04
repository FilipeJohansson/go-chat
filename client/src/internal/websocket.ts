import { Packet } from "../proto/packets"

type PacketHandler = (packet: Packet) => void
type VoidHandler = () => void

interface WebSocketClientOptions {
  protocols?: string[]
  onConnected?: VoidHandler
  onDisconnected?: VoidHandler
  onPacket?: PacketHandler
}

export class WebSocketClient {
  private static instance: WebSocketClient | null = null

  private socket: WebSocket | null = null
  private url: string = ""
  private onConnected?: VoidHandler
  private onDisconnected?: VoidHandler
  private onPacket?: PacketHandler

  constructor() {}

  static getInstance(): WebSocketClient {
    if (!this.instance) this.instance = new WebSocketClient()
    return this.instance
  }

  configure(options: WebSocketClientOptions) {
    this.onConnected = options.onConnected
    this.onDisconnected = options.onDisconnected
    this.onPacket = options.onPacket
  }

  connect(url: string) {
    if (
      this.socket &&
      (this.socket.readyState === WebSocket.OPEN ||
       this.socket.readyState === WebSocket.CONNECTING)
    ) {
      console.warn("WebSocket already connected or connecting.")
      return
    }

    this.url = url
    this.socket = new WebSocket(url)
    this.socket.binaryType = 'arraybuffer'

    this.socket.onopen = () => {
      this.onConnected?.()
    }
  
    this.socket.onclose = () => {
      this.onDisconnected?.()
    }
  
    this.socket.onerror = () => {
      this.onDisconnected?.()
    }

    this.socket.onmessage = (event) => {
      try {
        const buffer: ArrayBuffer = event.data as ArrayBuffer
        const data = new Uint8Array(buffer)

        const packet = Packet.decode(data)
        this.onPacket?.(packet)
      } catch (err) {
        console.error("Error decoding packet", err, event.data)
      }
    }
  }

  send(packet: Packet): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) return
    packet.senderId = 0

    const data: Uint8Array = Packet.encode(packet).finish()
    this.socket.send(data)
  }

  close(code: number = 1000, reason: string = "") {
    this.socket?.close(code, reason)
  }

  clear() {
    if (this.socket) {
      this.socket.onopen = null
      this.socket.onclose = null
      this.socket.onerror = null
      this.socket.onmessage = null
      this.socket.close()
      this.socket = null
    }
  
    this.onConnected = undefined
    this.onDisconnected = undefined
    this.onPacket = undefined
  }

  isOpen(): boolean {
    return this.socket?.readyState === WebSocket.OPEN
  }
}
