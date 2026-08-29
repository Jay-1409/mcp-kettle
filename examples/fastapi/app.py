from fastapi import FastAPI

app = FastAPI()


@app.get("/hello/{name}")
def hello(name: str, excited: bool = False):
    return {"message": f"Hello, {name}{'!' if excited else '.'}"}


@app.delete("/messages/{message_id}")
def delete_message(message_id: int):
    return {"deleted": message_id}
