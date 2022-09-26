package database

func GetAllPeerConnInfo() ([]*PeerConnInfo, error) {
	db := GetDatabase()
	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("peer_conn_info", "id")
	if err != nil {
		return nil, err
	}

	peerConnInfoList := []*PeerConnInfo{}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		peerConnInfo := obj.(*PeerConnInfo)
		peerConnInfoList = append(peerConnInfoList, peerConnInfo)
	}

	return peerConnInfoList, nil
}

func InsertPeerConnInfo(info *PeerConnInfo) error {
	db := GetDatabase()
	txn := db.Txn(true)
	err := txn.Insert("peer_conn_info", info)
	if err != nil {
		return err
	}
	txn.Commit()

	return nil
}
