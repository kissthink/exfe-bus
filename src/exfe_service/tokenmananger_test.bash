curl http://127.0.0.1:23333/TokenManager?method=Generate -d '{"resource":"abcde","expire_after_seconds":12}'
curl http://127.0.0.1:23333/TokenManager?method=Get -d 'deae3cee0be68e2ae2c590f0a1b5bb032168477d2d2c2a515b652042331b0220'
curl http://127.0.0.1:23333/TokenManager?method=Verify -d '{"token":"deae3cee0be68e2ae2c590f0a1b5bb032168477d2d2c2a515b652042331b0220","resource":"abcde"}'
curl http://127.0.0.1:23333/TokenManager?method=Refresh -d '{"token":"deae3cee0be68e2ae2c590f0a1b5bb032168477d2d2c2a515b652042331b0220","expire_after_seconds":-1}'
curl http://127.0.0.1:23333/TokenManager?method=Expire -d '"deae3cee0be68e2ae2c590f0a1b5bb032168477d2d2c2a515b652042331b0220"'
curl http://127.0.0.1:23333/TokenManager?method=Delete -d '"deae3cee0be68e2ae2c590f0a1b5bb032168477d2d2c2a515b652042331b0220"'