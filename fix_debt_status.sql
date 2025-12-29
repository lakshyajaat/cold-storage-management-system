-- Update old 'approved' debt requests to 'used' if there's a completed gate pass
-- for the same customer and thock number

UPDATE debt_requests dr
SET status = 'used',
    gate_pass_id = (
        SELECT gp.id 
        FROM gate_passes gp
        JOIN customers c ON gp.customer_id = c.id
        WHERE c.phone = dr.customer_phone 
        AND gp.thock_number = dr.thock_number
        AND gp.status = 'completed'
        ORDER BY gp.completed_at DESC
        LIMIT 1
    )
WHERE dr.status = 'approved'
AND EXISTS (
    SELECT 1 FROM gate_passes gp
    JOIN customers c ON gp.customer_id = c.id
    WHERE c.phone = dr.customer_phone 
    AND gp.thock_number = dr.thock_number
    AND gp.status = 'completed'
);
