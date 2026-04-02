class ComputeAnalytics:
    def __init__(self):
        pass

    def computeAnalytics(self, request: dict) -> dict:
        receipts = request.get("receipts")

        # compute the total spent across all the receipts
        totalAmount = sum(float(receipt.get("totalAmount")) for receipt in receipts)

        # category wise breakdown
        categoryTotals: dict[str, float] = {}

        for receipt in receipts:
            for item in receipt.get("items"):
                category = item.get("category")
                categoryTotals[category] = categoryTotals.get(category, 0) + item.get("price")

        categoryBreakdown = [
            { "category": cat, "amount": round(amt, 2) }
            for cat, amt in sorted(categoryTotals.items(), key=lambda x: x[1], reverse=True)
        ]

        # daily spending totals
        dailyTotals: dict[str, float] = {}
        for receipt in receipts:
            date = receipt.get("date", "")
            dailyTotals[date] = dailyTotals.get(date, 0) + receipt.get("totalAmount")

        dailySpending = [
            { "date": date, "amount": round(amt, 2) }
            for date, amt in sorted(dailyTotals.items())
        ]

        return {
            "totalAmount": round(totalAmount, 2),
            "categoryBreakdown": categoryBreakdown,
            "dailySpending": dailySpending
        }