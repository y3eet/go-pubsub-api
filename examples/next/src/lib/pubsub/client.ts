export class GoPubSubClient {
  constructor(url: string) {
    this.url = url;
  }
  private url: string;
  subscribe<T>(
    topic: string,
    onMessage: (message: T) => void,
    onError: (error: Event) => void,
  ): () => void {
    const ws = new WebSocket(`${this.url}/subscribe/${topic}`);

    ws.onmessage = (event) => {
      onMessage(JSON.parse(event.data) as T);
    };

    ws.onerror = (error) => {
      console.error("WebSocket failed:", error);
      ws.close();
      onError(error);
    };

    return () => {
      if (
        ws.readyState === WebSocket.OPEN ||
        ws.readyState === WebSocket.CONNECTING
      ) {
        ws.close();
      }
    };
  }

  async publish<T>(topic: T, message: T): Promise<void> {
    const res = await fetch(`${this.url}/publish/${topic}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(message),
    });

    if (!res.ok) {
      throw new Error(`Failed to publish message: ${res.statusText}`);
    }
  }
}
