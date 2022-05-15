package viewer

import (
	"regexp"
	"strings"
)

type CategoryData struct {
	Name        string
	DisplayName string
	Checked     bool
	Visibility  bool
	Open        bool
	Children    []*CategoryData
}

func NewCategoryData(categoryName, displayName string) *CategoryData {
	return &CategoryData{Name: categoryName, DisplayName: displayName, Checked: true, Visibility: true, Open: false, Children: []*CategoryData{}}
}

var logCategory = regexp.MustCompile(`^Log.+`)

const logCategoryParentKey = "Log*"

type CaregoryDataBuilder struct {
	parentCategories map[string]*CategoryData
	root             *CategoryData
}

func NewCaregoryDataBuilder() *CaregoryDataBuilder {
	return &CaregoryDataBuilder{}
}

func (c *CaregoryDataBuilder) hasParent(categoryName string) bool {
	return strings.ContainsRune(categoryName, '_')
}

func (c *CaregoryDataBuilder) getParentKey(categoryName string) string {
	if logCategory.MatchString(categoryName) {
		if categoryName == logCategoryParentKey {
			return ""
		} else {
			// Logから始まるカテゴリが多いためLog*の子要素としてViewerのカテゴリ一覧の可読性を上げる
			return logCategoryParentKey
		}
	} else if strings.ContainsRune(categoryName, '_') {
		s := strings.Split(categoryName, "_")
		return strings.Join(s[:len(s)-1], "_")
	} else {
		return ""
	}
}

func (c *CaregoryDataBuilder) getParentCategoryData(parentKey string) *CategoryData {
	if data, ok := c.parentCategories[parentKey]; ok {
		return data
	} else {
		var data *CategoryData
		if parentParentKey := c.getParentKey(parentKey); parentParentKey != "" {
			parentParent := c.getParentCategoryData(parentParentKey)
			displayName := parentKey[len(parentParentKey)+1:]
			data = NewCategoryData(displayName, displayName)
			parentParent.Children = append(parentParent.Children, data)
		} else {
			displayName := parentKey
			data = NewCategoryData(displayName, displayName)
			c.root.Children = append(c.root.Children, data)
		}
		c.parentCategories[parentKey] = data
		return data
	}
}

func (c *CaregoryDataBuilder) CreateCategoryData(categoryNames []string) *CategoryData {
	c.parentCategories = make(map[string]*CategoryData)

	c.root = NewCategoryData("All", "All")
	c.root.Open = true // 1階層目だけ最初から開けておく
	c.parentCategories["All"] = c.root

	for _, categoryName := range categoryNames {
		parentKey := c.getParentKey(categoryName)
		if parentKey != "" {
			parentCategoryData := c.getParentCategoryData(parentKey)
			displayName := categoryName
			if parentKey != logCategoryParentKey {
				displayName = categoryName[len(parentKey)+1:]
			}
			categoryData := NewCategoryData(categoryName, displayName)
			parentCategoryData.Children = append(parentCategoryData.Children, categoryData)
		} else {
			categoryData := NewCategoryData(categoryName, categoryName)
			c.root.Children = append(c.root.Children, categoryData)
		}
	}

	return c.root
}
