package db

type DBObject interface {
	*Community | *CommunityJson | *CommunitySettings | *CommunityAssociate | *CommunityToken | *CommunityFaq | *CommunityExtras | *Collection | *CollectionSettings | *CollectionPrefix | *CollectionAddress | *CollectionArtist | *CollectionAsset | *CollectionExcludedAsset | *CollectionExtras | *Property | *PropertyValue | *PropertyValueExtras | *Provider | *ProviderAddress
}
