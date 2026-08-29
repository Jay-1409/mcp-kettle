from fastapi import APIRouter

router = APIRouter()


@router.get("/projects")
def list_projects(offset: int = 0, limit: int = 20):
    return {"offset": offset, "limit": limit, "projects": []}


@router.get("/projects/{project_id}")
def get_project(project_id: int):
    return {"project_id": project_id}


@router.post("/projects")
def create_project(name: str):
    return {"name": name}


@router.put("/projects/{project_id}")
def update_project(project_id: int, name: str):
    return {"project_id": project_id, "name": name}


@router.delete("/projects/{project_id}")
def delete_project(project_id: int):
    return {"deleted": project_id}
