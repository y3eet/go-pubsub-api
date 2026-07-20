import asyncio
import websockets

topic = "my-topic"
WS_URL = f"ws://localhost:8080/subscribe/{topic}"


async def connect():
    while True:
        try:
            async with websockets.connect(WS_URL) as websocket:
                print(f"Connected to {WS_URL}")

                while True:
                    message = await websocket.recv()
                    print(f"Received: {message}")

        except Exception as e:
            print(f"Disconnected: {e}")

        print("Reconnecting in 5 seconds...\n")
        await asyncio.sleep(5)


asyncio.run(connect())
