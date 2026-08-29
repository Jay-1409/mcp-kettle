from fastapi import APIRouter

router = APIRouter()


@router.get("/billing/plans")
def list_plans():
    return {"plans": []}


@router.get("/billing/invoices")
def list_invoices(limit: int = 20):
    return {"limit": limit, "invoices": []}


@router.get("/billing/invoices/{invoice_id}")
def get_invoice(invoice_id: int):
    return {"invoice_id": invoice_id}


@router.post("/billing/invoices/{invoice_id}/pay")
def pay_invoice(invoice_id: int):
    return {"invoice_id": invoice_id, "paid": True}
