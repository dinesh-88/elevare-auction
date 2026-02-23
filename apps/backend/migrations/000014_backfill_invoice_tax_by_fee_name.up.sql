UPDATE invoices i
SET tax_total_minor = COALESCE((
  SELECT SUM((fee ->> 'amount')::bigint)
  FROM jsonb_array_elements(COALESCE(i.snapshot -> 'fees', '[]'::jsonb)) fee
  WHERE COALESCE((fee ->> 'taxable')::boolean, false)
     OR UPPER(COALESCE(fee ->> 'name', '')) LIKE '%VAT%'
     OR UPPER(COALESCE(fee ->> 'name', '')) LIKE '%TAX%'
     OR UPPER(COALESCE(fee ->> 'name', '')) LIKE '%GST%'
), 0)
WHERE i.snapshot IS NOT NULL;
