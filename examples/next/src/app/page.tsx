"use client";

import { pubsub } from "@/lib/pubsub";
import { useEffect, useState } from "react";

type ChatMessage = { message: string };

const ROOMS = ["general", "random", "tech"] as const;

export default function Home() {
  const [room, setRoom] = useState<string>(ROOMS[0]);
  const [messagesByRoom, setMessagesByRoom] = useState<
    Record<string, ChatMessage[]>
  >(
    () =>
      Object.fromEntries(ROOMS.map((name) => [name, []])) as Record<
        string,
        ChatMessage[]
      >,
  );

  useEffect(() => {
    const unsubscribe = pubsub.subscribe<ChatMessage>(
      `chat.${room}`,
      (message) => {
        console.log("Received message:", message);
        setMessagesByRoom((prev) => ({
          ...prev,
          [room]: [...(prev[room] ?? []), message],
        }));
      },
      (error) => {
        console.error("Subscription error:", error);
      },
    );

    return unsubscribe;
  }, [room]);

  const currentMessages = messagesByRoom[room] ?? [];

  return (
    <div className="flex min-h-screen w-full justify-center bg-zinc-100 p-4">
      <main className="flex w-full max-w-3xl flex-col gap-6 rounded-xl bg-white p-6 shadow-sm sm:p-8">
        <section className="space-y-2">
          <h1 className="text-2xl font-semibold text-zinc-900">Rooms</h1>
          <div className="flex flex-wrap gap-2">
            {ROOMS.map((roomName) => (
              <button
                key={roomName}
                type="button"
                onClick={() => setRoom(roomName)}
                className={`rounded-full border px-4 py-1.5 text-sm transition-colors ${
                  room === roomName
                    ? "border-zinc-900 bg-zinc-900 text-white"
                    : "border-zinc-300 bg-white text-zinc-700 hover:border-zinc-400"
                }`}
              >
                #{roomName}
              </button>
            ))}
          </div>
        </section>
        <div>
          <input
            type="text"
            id="message-input"
            placeholder={`Message #${room}...`}
            className="w-full rounded-md border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-800 focus:border-zinc-500 focus:outline-none"
            onKeyDown={async (e) => {
              if (e.key === "Enter") {
                const target = e.target as HTMLInputElement;
                const message = target.value.trim();
                if (message) {
                  try {
                    await pubsub.publish(`chat.${room}`, { message });
                    target.value = "";
                  } catch (error) {
                    console.error("Failed to publish message:", error);
                  }
                }
              }
            }}
          />
          <button
            type="button"
            onClick={async () => {
              const input = document.getElementById(
                "message-input",
              ) as HTMLInputElement | null;
              if (!input) return;

              const message = input.value.trim();
              if (message) {
                try {
                  await pubsub.publish(`chat.${room}`, { message });
                  input.value = "";
                } catch (error) {
                  console.error("Failed to publish message:", error);
                }
              }
            }}
            className="mt-2 rounded-md bg-zinc-900 px-4 py-2 text-sm text-white hover:bg-zinc-800"
          >
            Send Message
          </button>
        </div>
        <section className="space-y-3">
          <h2 className="text-lg font-medium text-zinc-800">
            Messages in #{room}
          </h2>
          <div className="max-h-105 overflow-y-auto rounded-lg border border-zinc-200 bg-zinc-50 p-3">
            {currentMessages.length === 0 ? (
              <p className="text-sm text-zinc-500">
                No messages yet for this room.
              </p>
            ) : (
              <ul className="space-y-2">
                {currentMessages.map((item, index) => (
                  <li
                    key={`${room}-${index}`}
                    className="rounded-md border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-800"
                  >
                    {item.message}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>
      </main>
    </div>
  );
}
