DO $$
DECLARE
    v_table_count INTEGER;
    v_role_count INTEGER;
    v_scope_count INTEGER;
    v_test_user_id UUID;
    v_test_event_id UUID;
    v_test_ledger_id UUID;
    v_test_tx_id UUID := gen_random_uuid();
    v_error_caught BOOLEAN := FALSE;
BEGIN
    RAISE NOTICE '====================================================';
    RAISE NOTICE 'STARTING DCISP DATABASE INFRASTRUCTURE VERIFICATION';
    RAISE NOTICE '====================================================';

    -- 1. VERIFY TABLES EXISTENCE
    SELECT COUNT(*) INTO v_table_count 
    FROM information_schema.tables 
    WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
    
    RAISE NOTICE '1. Total tables created: % (Expected >= 35)', v_table_count;
    IF v_table_count < 35 THEN
        RAISE EXCEPTION 'Verification Failed: Table count is less than 35';
    END IF;

    -- 2. VERIFY SEED DATA
    SELECT COUNT(*) INTO v_role_count FROM roles;
    SELECT COUNT(*) INTO v_scope_count FROM scopes;
    RAISE NOTICE '2. Total roles seeded: % (Expected 10)', v_role_count;
    RAISE NOTICE '   Total scopes seeded: % (Expected 8)', v_scope_count;
    
    IF v_role_count < 10 OR v_scope_count < 8 THEN
        RAISE EXCEPTION 'Verification Failed: Seed data for roles or scopes is incomplete';
    END IF;

    -- 3. TEST IMMUTABILITY TRIGGER ON attendance_event_logs (BR-025)
    INSERT INTO users (id, email, password_hash, full_name, status)
    VALUES (gen_random_uuid(), 'test_verifier@dcisp.internal', 'hash123', 'Audit Verifier User', 'ACTIVE')
    RETURNING id INTO v_test_user_id;

    INSERT INTO attendance_event_logs (id, user_id, event_type, timestamp, method)
    VALUES (gen_random_uuid(), v_test_user_id, 'CHECK_IN', CURRENT_TIMESTAMP, 'NFC')
    RETURNING id INTO v_test_event_id;

    -- Try UPDATE on attendance_event_logs
    v_error_caught := FALSE;
    BEGIN
        UPDATE attendance_event_logs SET event_type = 'CHECK_OUT' WHERE id = v_test_event_id;
    EXCEPTION WHEN OTHERS THEN
        v_error_caught := TRUE;
        RAISE NOTICE '3a. Success: Trigger correctly BLOCKED direct UPDATE on attendance_event_logs';
    END;
    IF v_error_caught = FALSE THEN
        RAISE EXCEPTION 'Verification Failed: Immutability trigger failed to block UPDATE on attendance_event_logs';
    END IF;

    -- Try DELETE on attendance_event_logs
    v_error_caught := FALSE;
    BEGIN
        DELETE FROM attendance_event_logs WHERE id = v_test_event_id;
    EXCEPTION WHEN OTHERS THEN
        v_error_caught := TRUE;
        RAISE NOTICE '3b. Success: Trigger correctly BLOCKED direct DELETE on attendance_event_logs';
    END;
    IF v_error_caught = FALSE THEN
        RAISE EXCEPTION 'Verification Failed: Immutability trigger failed to block DELETE on attendance_event_logs';
    END IF;

    -- 4. TEST IMMUTABILITY TRIGGER ON financial_ledgers (BR-023)
    INSERT INTO financial_ledgers (id, transaction_id, account_code, direction, amount, reference_table, reference_id, narration)
    VALUES (gen_random_uuid(), v_test_tx_id, '1001-CASH', 'DEBIT', 500000.00, 'projects', gen_random_uuid(), 'Verification Entry')
    RETURNING id INTO v_test_ledger_id;

    -- Try UPDATE on financial_ledgers
    v_error_caught := FALSE;
    BEGIN
        UPDATE financial_ledgers SET amount = 600000.00 WHERE id = v_test_ledger_id;
    EXCEPTION WHEN OTHERS THEN
        v_error_caught := TRUE;
        RAISE NOTICE '4a. Success: Trigger correctly BLOCKED direct UPDATE on financial_ledgers';
    END;
    IF v_error_caught = FALSE THEN
        RAISE EXCEPTION 'Verification Failed: Immutability trigger failed to block UPDATE on financial_ledgers';
    END IF;

    -- Try DELETE on financial_ledgers
    v_error_caught := FALSE;
    BEGIN
        DELETE FROM financial_ledgers WHERE id = v_test_ledger_id;
    EXCEPTION WHEN OTHERS THEN
        v_error_caught := TRUE;
        RAISE NOTICE '4b. Success: Trigger correctly BLOCKED direct DELETE on financial_ledgers';
    END;
    IF v_error_caught = FALSE THEN
        RAISE EXCEPTION 'Verification Failed: Immutability trigger failed to block DELETE on financial_ledgers';
    END IF;

    RAISE NOTICE '====================================================';
    RAISE NOTICE 'ALL DATABASE INFRASTRUCTURE CHECKS PASSED 100%%!';
    RAISE NOTICE '====================================================';
END $$;
