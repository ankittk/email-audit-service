#!/bin/bash

BASE_URL="http://localhost:8080"

echo "🧪 Testing Email Audit Service"
echo "================================"

# Test health check
echo "1. Testing health check..."
curl -X GET "$BASE_URL/health" | jq

echo -e "\n2. Testing email upload..."

curl -X POST "$BASE_URL/upload-email" \
  -F "file=@test/sample.eml" \
  -F "user_id=user123" \
  -F "company_id=company456" | jq

echo -e "\n3. Testing rules listing..."
curl -X GET "$BASE_URL/rules?company_id=company456" | jq

echo -e "\n4. Testing audit summary..."
curl -X GET "$BASE_URL/audit/summary?company_id=company456&page=1&page_size=10" | jq
```

### Sample .eml file (test/sample.eml)

```
From: john.doe@company.com
To: client@example.com
Subject: Project Update
Date: Tue, 17 Jun 2025 10:30:00 +0000
Content-Type: text/plain; charset=UTF-8

Dear Mr. Smith,

I hope this email finds you well. I wanted to provide you with an update on our project progress.

We have successfully completed the first phase of the project and are pleased to report that we are on schedule. The team has been working diligently to ensure all deliverables meet your specifications.

Please let me know if you have any questions or concerns. I would be happy to schedule a call to discuss the next steps.

Thank you for your continued trust in our services.

Best regards,
John Doe
Project Manager
