UPDATE invoices i
SET tax_total_minor = COALESCE((
  SELECT SUM((fee ->> 'amount')::bigint)
  FROM jsonb_array_elements(COALESCE(i.snapshot -> 'fees', '[]'::jsonb)) fee
  WHERE COALESCE((fee ->> 'taxable')::boolean, false)
), 0)
WHERE i.snapshot IS NOT NULL;
