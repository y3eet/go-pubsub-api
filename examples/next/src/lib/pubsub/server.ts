export class GoPubSubServer {
  constructor(url: string) {
    this.url = url;
  }
  private url: string;

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
