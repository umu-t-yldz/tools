import json
import pandas as pd
from openpyxl import load_workbook
from openpyxl.styles import PatternFill

# JSON dosyasını oku
with open("result.json", "r", encoding="utf-8") as f:
    data = json.load(f)

# JSON verisinden status, category ve target alanlarını çek
rows = []
for entry in data:
    rows.append({
        "status": entry.get("status", ""),
        "category": entry.get("category", ""),
        "target": entry.get("target", "")
    })

# DataFrame oluştur
df = pd.DataFrame(rows, columns=["status", "category", "target"])

# Özel sıralama için bir sütun oluştur
def sort_order(row):
    order = {
        ("connected", "domain"): 0,
        ("connected", "url"): 1,
        ("not_connected", "domain"): 2,
        ("not_connected", "url"): 3
    }
    return order.get((row["status"], row["category"]), 4)

df["sort_key"] = df.apply(sort_order, axis=1)

# Sıralamayı uygula
df = df.sort_values("sort_key").drop(columns=["sort_key"])

# Excel'e yaz
excel_file = "ndba_output.xlsx"
df.to_excel(excel_file, index=False)

# Excel dosyasını aç
wb = load_workbook(excel_file)
ws = wb.active

# Soft renkler (A sütunu için)
soft_red = PatternFill(start_color="FFC7CE", end_color="FFC7CE", fill_type="solid")
soft_green = PatternFill(start_color="C6EFCE", end_color="C6EFCE", fill_type="solid")

# A sütununu renklendir
for row in range(2, ws.max_row + 1):  # 1. satır başlık olduğu için 2’den başlıyoruz
    cell = ws[f"A{row}"]
    if cell.value == "connected":
        cell.fill = soft_red
    elif cell.value == "not_connected":
        cell.fill = soft_green

# Dosyayı kaydet
wb.save(excel_file)

print(excel_file, " oluşturuldu.")
