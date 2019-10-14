import java.util.List;
import java.util.ArrayList;

import com.mongodb.client.MongoCollection;
import org.bson.Document;
import org.bson.conversions.Bson;
import org.bson.types.ObjectId;

import static com.mongodb.client.model.Filters.eq;
public class LinkRepository {

    //private final List<Link> links;
    private final MongoCollection<Document> links;

    /*public LinkRepository() {
        //links = new ArrayList<>();
        //add some links to start off with
        //links.add(new Link("http://howtographql.com", "Your favorite GraphQL page"));
        //links.add(new Link("http://graphql.org/learn/", "The official docks"));
    }*/

    public LinkRepository(MongoCollection<Document> links) {
        this.links = links;
    }

    public Link findById(String id) {
        Document doc = links.find(eq("_id", new ObjectId(id))).first();
        return link(doc);
    }

    public List<Link> getAllLinks() {
        //return links;
        List<Link> allLinks = new ArrayList<>();
        for (Document doc : links.find()) {
            allLinks.add(link(doc));
        }
        return allLinks;
    }

    public void saveLink(Link link) {
        //links.add(link);
        Document doc = new Document();
        doc.append("url", link.getUrl());
        doc.append("description", link.getDescription());
        links.insertOne(doc);
    }

    // link(doc)
    private Link link(Document doc) {
        return new Link(
                doc.get("_id").toString(),
                doc.getString("url"),
                doc.getString("description"));
    }
}