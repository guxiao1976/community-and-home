package contact

import communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"

func categoryToString(c communityv1.ContactCategory) string {
	switch c {
	case communityv1.ContactCategory_CONTACT_CATEGORY_WATER:
		return "water"
	case communityv1.ContactCategory_CONTACT_CATEGORY_ELECTRICITY:
		return "electricity"
	case communityv1.ContactCategory_CONTACT_CATEGORY_GAS:
		return "gas"
	case communityv1.ContactCategory_CONTACT_CATEGORY_UNICOM:
		return "unicom"
	case communityv1.ContactCategory_CONTACT_CATEGORY_MOBILE:
		return "mobile"
	case communityv1.ContactCategory_CONTACT_CATEGORY_TELECOM:
		return "telecom"
	case communityv1.ContactCategory_CONTACT_CATEGORY_POLICE:
		return "police"
	default:
		return ""
	}
}

func stringToCategory(s string) communityv1.ContactCategory {
	switch s {
	case "water":
		return communityv1.ContactCategory_CONTACT_CATEGORY_WATER
	case "electricity":
		return communityv1.ContactCategory_CONTACT_CATEGORY_ELECTRICITY
	case "gas":
		return communityv1.ContactCategory_CONTACT_CATEGORY_GAS
	case "unicom":
		return communityv1.ContactCategory_CONTACT_CATEGORY_UNICOM
	case "mobile":
		return communityv1.ContactCategory_CONTACT_CATEGORY_MOBILE
	case "telecom":
		return communityv1.ContactCategory_CONTACT_CATEGORY_TELECOM
	case "police":
		return communityv1.ContactCategory_CONTACT_CATEGORY_POLICE
	default:
		return communityv1.ContactCategory_CONTACT_CATEGORY_UNSPECIFIED
	}
}
