package db

import (
	"fmt"
)

func getTable(obj interface{}) string {
	switch obj.(type) {
	case Community, *Community:
		return fmt.Sprintf("%s.community", arc53Database())
	case CommunityJson, *CommunityJson:
		return fmt.Sprintf("%s.community_json", arc53Database())
	case CommunitySettings, *CommunitySettings:
		return fmt.Sprintf("%s.community_settings", arc53Database())
	case CommunityToken, *CommunityToken:
		return fmt.Sprintf("%s.community_token", arc53Database())
	case CommunityAssociate, *CommunityAssociate:
		return fmt.Sprintf("%s.community_associate", arc53Database())
	case CommunityFaq, *CommunityFaq:
		return fmt.Sprintf("%s.community_faq", arc53Database())
	case CommunityExtras, *CommunityExtras:
		return fmt.Sprintf("%s.community_extras", arc53Database())
	case Collection, *Collection:
		return fmt.Sprintf("%s.collection", arc53Database())
	case CollectionSettings, *CollectionSettings:
		return fmt.Sprintf("%s.collection_settings", arc53Database())
	case CollectionPrefix, *CollectionPrefix:
		return fmt.Sprintf("%s.collection_prefix", arc53Database())
	case CollectionAddress, *CollectionAddress:
		return fmt.Sprintf("%s.collection_address", arc53Database())
	case CollectionAsset, *CollectionAsset:
		return fmt.Sprintf("%s.collection_asset", arc53Database())
	case CollectionExcludedAsset, *CollectionExcludedAsset:
		return fmt.Sprintf("%s.collection_excluded_asset", arc53Database())
	case CollectionArtist, *CollectionArtist:
		return fmt.Sprintf("%s.collection_artist", arc53Database())
	case CollectionExtras, *CollectionExtras:
		return fmt.Sprintf("%s.collection_extras", arc53Database())
	case Property, *Property:
		return fmt.Sprintf("%s.property", arc53Database())
	case PropertyValue, *PropertyValue:
		return fmt.Sprintf("%s.property_value", arc53Database())
	case PropertyValueExtras, *PropertyValueExtras:
		return fmt.Sprintf("%s.property_value_extras", arc53Database())
	case Provider, *Provider:
		return fmt.Sprintf("%s.provider", arc53Database())
	case ProviderAddress, *ProviderAddress:
		return fmt.Sprintf("%s.provider_address", arc53Database())
	default:
		return ""
	}
}
