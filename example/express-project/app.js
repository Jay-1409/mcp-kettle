const express = require("express");

const app = express();

app.get("/health", (_request, response) => response.json({ ok: true }));
app.get("/users", (_request, response) => response.json([]));
app.get("/users/:userId", (_request, response) => response.json({}));
app.post("/users", (_request, response) => response.status(201).json({}));
app.delete("/users/:userId", (_request, response) => response.status(204).end());
app.get("/projects", (_request, response) => response.json([]));
app.get("/projects/:projectId", (_request, response) => response.json({}));
app.post("/projects", (_request, response) => response.status(201).json({}));

app.listen(3000);
