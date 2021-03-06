package parser

import (
	"encoding/json"
	"errors"
)

/*
func ParseAuction(data *RawAuctionData) (r *Auction) {
	r = new(Auction)
	r.Auc = (*data)["auc"].(uint64)
	r.Item = (*data)["item"].(uint64)
	r.Owner = (*data)["owner"].(string)
	r.Realm = (*data)["ownerRealm"].(string)
	r.Bid = (*data)["bid"].(uint64)
	r.Buyout = (*data)["buyout"].(uint64)
	r.Quantity = (*data)["quantity"].(uint32)
	r.TimeLeft = (*data)["timeLeft"].(string)
	r.Rand = (*data)["rand"].(int64)
	r.Seed = (*data)["seed"].(int64)
	r.Context = (*data)["context"].(int64)
	return
}
*/
var MalformedBlob error = errors.New("Blob is malformed")

func ParseSnapshot(data []byte) (snapshot *SnapshotData, err error) {
	snapshot = new(SnapshotData)
	err = json.Unmarshal(data, snapshot)
	if err != nil {
		return nil, err
	}
	if snapshot.Realms == nil {
		return nil, MalformedBlob
	}
	if len(snapshot.Realms) == 0 {
		return nil, MalformedBlob
	}
	if snapshot.Auctions == nil {
		return nil, MalformedBlob
	}
	return snapshot, nil
}

func MakeBaseAuction(auc *Auction) (bse *BaseAuction) {
	bse = new(BaseAuction)
	*bse = auc.BaseAuction
	return
}

func MakeAuctionWithBonus(auc *Auction) (bns *AuctionWithBonus) {
	bns = new(AuctionWithBonus)
	bns.BaseAuction = auc.BaseAuction
	bns.BonusPart = auc.BonusPart
	return
}

func MakeAuctionWithMods(auc *Auction) (mod *AuctionWithMods) {
	mod = new(AuctionWithMods)
	mod.BaseAuction = auc.BaseAuction
	mod.ModsPart = auc.ModsPart
	mod.BonusPart = auc.BonusPart
	return
}

func MakePetAuction(auc *Auction) (pet *PetAuction) {
	pet = new(PetAuction)
	pet.BaseAuction = auc.BaseAuction
	pet.ModsPart = auc.ModsPart
	pet.PetPart = auc.PetPart
	return
}

func PackAuctionData(auc *Auction) (blob []byte) {
	switch {
	case auc.PetSpeciesId != 0:
		blob, _ = json.Marshal(MakePetAuction(auc))
	case auc.Modifiers != nil:
		blob, _ = json.Marshal(MakeAuctionWithMods(auc))
	case auc.BonusLists != nil:
		blob, _ = json.Marshal(MakeAuctionWithBonus(auc))
	default:
		blob, _ = json.Marshal(MakeBaseAuction(auc))
	}
	return
}
